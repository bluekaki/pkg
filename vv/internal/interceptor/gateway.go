package interceptor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/vv/internal/protos/gen"
	"github.com/bluekaki/pkg/vv/options"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

var gwHeader = struct {
	key   string
	value string
}{
	key:   "grpc-gateway",
	value: "bluekaki/grpcgw/m/v1.1",
}

// ForwardedByGrpcGateway whether forwarded by grpc gateway
func ForwardedByGrpcGateway(ctx context.Context) bool {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}

	return forwardedByGrpcGateway(meta)
}

func forwardedByGrpcGateway(meta metadata.MD) bool {
	values := meta.Get(gwHeader.key)
	if len(values) == 0 {
		return false
	}

	return values[0] == gwHeader.value
}

// NewGatewayInterceptor create a gateway interceptor
func NewGatewayInterceptor(logger *zap.Logger, metrics func(http.Handler), notify NotifyHandler) *GatewayInterceptor {
	if metrics != nil {
		metrics(promhttp.Handler())
	}

	return &GatewayInterceptor{
		logger:  logger,
		metrics: metrics,
		notify:  notify,
	}
}

// GatewayInterceptor the gateway's interceptor
type GatewayInterceptor struct {
	logger  *zap.Logger
	metrics func(http.Handler)
	notify  NotifyHandler
}

// UnaryInterceptor a interceptor for gateway unary operations
func (g *GatewayInterceptor) UnaryInterceptor(ctx context.Context, fullMethod string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	ts := time.Now()
	meta, _ := metadata.FromOutgoingContext(ctx)
	meta.Set(gwHeader.key, gwHeader.value)

	journalID := meta.Get(JournalID)[0]

	doJournal := false
	if proto.GetExtension(FileDescriptor.Options(fullMethod), options.E_Journal).(bool) {
		doJournal = true
	}

	defer func() { // double recover for safety
		if p := recover(); p != nil {
			err = errors.Panic(p)
			errVerbose := fmt.Sprintf("got double panic => error: %+v", err)
			if g.notify != nil {
				g.notify((&errors.AlertMessage{
					JournalId:    journalID,
					ErrorVerbose: errVerbose,
				}).Init())
			}

			err = status.New(codes.Internal, "got double panic").Err()
			g.logger.Error(fmt.Sprintf("%s %s", journalID, errVerbose))
		}
	}()

	defer func() {
		if p := recover(); p != nil {
			err = errors.Panic(p)
			errVerbose := fmt.Sprintf("got panic => error: %+v", err)
			if g.notify != nil {
				g.notify((&errors.AlertMessage{
					JournalId:    journalID,
					ErrorVerbose: errVerbose,
				}).Init())
			}

			s, _ := status.New(codes.Internal, "got panic").WithDetails(&pb.Stack{Verbose: errVerbose})
			err = s.Err()
		}

		if doJournal {
			journal := &pb.Journal{
				Id: journalID,
				Request: &pb.Request{
					Restapi: true,
					Method:  fullMethod,
					Metadata: func() map[string]string {
						mp := make(map[string]string)
						for key, values := range meta {
							if key == URI {
								mp[key] = QueryUnescape(values[0])
								continue
							}

							if toLoggedMetadata[key] {
								mp[key] = values[0]
							}
						}
						return mp
					}(),
					Payload: func() *anypb.Any {
						if req == nil {
							return nil
						}

						any, _ := anypb.New(req.(proto.Message))
						return any
					}(),
				},
				Response: &pb.Response{
					Code: codes.OK.String(),
					Payload: func() *anypb.Any {
						if reply == nil {
							return nil
						}

						any, _ := anypb.New(reply.(proto.Message))
						return any
					}(),
				},
				Success: err == nil,
			}

			if err != nil {
				s, _ := status.FromError(err)
				journal.Response.Code = s.Code().String()
				journal.Response.Message = s.Message()

				if len(s.Details()) > 0 {
					journal.Response.ErrorVerbose = s.Details()[0].(*pb.Stack).Verbose
				}
				err = status.New(s.Code(), s.Message()).Err() // reset detail
			}

			journal.CostSeconds = time.Since(ts).Seconds()

			if err == nil {
				g.logger.Info("gateway unary interceptor", zap.Any("journal", marshalJournal(journal)))

			} else {
				g.logger.Error("gateway unary interceptor", zap.Any("journal", marshalJournal(journal)))
			}

			if g.metrics != nil {
				method := fullMethod

				if http := proto.GetExtension(FileDescriptor.Options(fullMethod), annotations.E_Http).(*annotations.HttpRule); http != nil {
					if x, ok := http.GetPattern().(*annotations.HttpRule_Get); ok {
						method = "GET " + x.Get
					} else if x, ok := http.GetPattern().(*annotations.HttpRule_Put); ok {
						method = "PUT " + x.Put
					} else if x, ok := http.GetPattern().(*annotations.HttpRule_Post); ok {
						method = "POST " + x.Post
					} else if x, ok := http.GetPattern().(*annotations.HttpRule_Delete); ok {
						method = "DELETE " + x.Delete
					} else if x, ok := http.GetPattern().(*annotations.HttpRule_Patch); ok {
						method = "PATCH " + x.Patch
					}
				}

				if alias := proto.GetExtension(FileDescriptor.Options(method), options.E_MetricsAlias).(string); alias != "" {
					method = alias
				}

				success := strconv.FormatBool(err == nil)
				requestCounter.WithLabelValues("gateway", method, success).Inc()
				requestDuration.WithLabelValues("gateway", method, success).Observe(time.Since(ts).Seconds())
			}
		}
	}()

	serviceName := strings.Split(fullMethod, "/")[1]

	var whitelistingValidator WhitelistingHandler
	if option := proto.GetExtension(FileDescriptor.Options(serviceName), options.E_Whitelisting).(*options.Handler); option != nil {
		whitelistingValidator = Validator.WhitelistingValidator(option.Name)
	}
	if option := proto.GetExtension(FileDescriptor.Options(fullMethod), options.E_MethodWhitelisting).(*options.Handler); option != nil {
		whitelistingValidator = Validator.WhitelistingValidator(option.Name)
	}

	if whitelistingValidator != nil {
		ok, err := whitelistingValidator(meta.Get(XForwardedFor)[0])
		if err != nil {
			s := status.New(codes.Aborted, codes.Aborted.String())
			s, _ = s.WithDetails(&pb.Stack{Verbose: fmt.Sprintf("%+v", err)})
			return s.Err()
		}
		if !ok {
			return status.Error(codes.Aborted, "ip does not allow access")
		}
	}

	err = invoker(metadata.NewOutgoingContext(ctx, meta), fullMethod, req, reply, cc, opts...)
	if err != nil {
		s, _ := status.FromError(err)
		if s.Code() == codes.Unavailable && g.notify != nil {
			g.notify((&errors.AlertMessage{
				JournalId:    journalID,
				ErrorVerbose: s.Proto().String(),
			}).Init())
		}
	}

	return
}
