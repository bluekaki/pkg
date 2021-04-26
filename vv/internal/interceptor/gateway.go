package interceptor

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/byepichi/pkg/pbutil"
	"github.com/byepichi/pkg/vv/internal/protos/gen"
	"github.com/byepichi/pkg/vv/options"

	"go.uber.org/zap"
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
	value: "koketama/grpcgw/m/v1.0",
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
func NewGatewayInterceptor(logger *zap.Logger, marshalJournal bool, notify notifyHandler) *GatewayInterceptor {
	return &GatewayInterceptor{
		logger:         logger,
		marshalJournal: marshalJournal,
		notify:         notify,
	}
}

// GatewayInterceptor the gateway's interceptor
type GatewayInterceptor struct {
	logger         *zap.Logger
	marshalJournal bool
	notify         notifyHandler
}

// UnaryInterceptor a interceptor for gateway unary operations
func (g *GatewayInterceptor) UnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	ts := time.Now()
	meta, _ := metadata.FromOutgoingContext(ctx)
	meta.Set(gwHeader.key, gwHeader.value)

	journalID := meta.Get(JournalID)[0]

	doJournal := false
	if proto.GetExtension(FileDescriptor.Options(method), options.E_Journal).(bool) {
		doJournal = true
	}

	defer func() { // double recover for safety
		if p := recover(); p != nil {
			err = status.New(codes.Internal, "got double panic").Err()
			g.logger.Error(fmt.Sprintf("got double panic => error: %+v\n%s", p, string(debug.Stack())))
		}
	}()

	defer func() {
		if p := recover(); p != nil {
			msg := fmt.Sprintf("got panic => error: %+v", p)
			info := string(debug.Stack())
			if g.notify != nil {
				g.notify("got panic", msg, info, journalID)
			}

			s, _ := status.New(codes.Internal, msg).WithDetails(&pb.Stack{Info: info})
			err = s.Err()
		}

		if doJournal {
			journal := &pb.Journal{
				Id: journalID,
				Request: &pb.Request{
					Restapi: true,
					Method:  method,
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

				journal.Response.Details = make([]*anypb.Any, len(s.Details()))
				for i, detail := range s.Details() {
					journal.Response.Details[i], _ = anypb.New(detail.(proto.Message))
				}

				err = status.New(s.Code(), s.Message()).Err() // reset detail
			}

			journal.CostSeconds = time.Since(ts).Seconds()

			var json interface{}
			if g.marshalJournal {
				json, _ = pbutil.ProtoMessage2JSON(journal)
			} else {
				json, _ = pbutil.ProtoMessage2Map(journal)
			}

			if err == nil {
				g.logger.Info("gateway unary interceptor", zap.Any("journal", json))
			} else {
				g.logger.Error("gateway unary interceptor", zap.Any("journal", json))
			}
		}
	}()

	serviceName := strings.Split(method, "/")[1]
	if option := proto.GetExtension(FileDescriptor.Options(serviceName), options.E_Whitelisting).(*options.Handler); option != nil {
		handler := Validator.WhitelistingValidator(option.Name)
		ok, err := handler(meta.Get(XForwardedFor)[0])
		if err != nil {
			return status.Error(codes.PermissionDenied, fmt.Sprintf("%+v", err))
		}
		if !ok {
			return status.Error(codes.PermissionDenied, "ip does not allow access")
		}
	}

	err = invoker(metadata.NewOutgoingContext(ctx, meta), method, req, reply, cc, opts...)
	if err != nil {
		s, _ := status.FromError(err)
		if s.Code() == codes.Unavailable && g.notify != nil {
			g.notify(codes.Unavailable.String(), "", s.Proto().String(), journalID)
		}
	}

	return
}
