package interceptor

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/vv/internal/pb"
	"github.com/bluekaki/pkg/vv/pkg/adapter"

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

func GatewayUnaryClientInterceptor(logger *zap.Logger, notify adapter.NotifyHandler, metrics func(http.Handler), projectName string) grpc.UnaryClientInterceptor {
	if metrics != nil {
		metrics(promhttp.Handler())
	}

	return func(ctx context.Context, fullMethod string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		ts := time.Now()

		meta, _ := metadata.FromOutgoingContext(ctx)
		meta.Set(gwHeader.key, gwHeader.value)

		journalID := meta.Get(JournalID)[0]

		method := fullMethod
		if httpRule, ok := getHttpRule(fullMethod); ok {
			if x, ok := httpRule.GetPattern().(*annotations.HttpRule_Get); ok {
				method = "GET " + x.Get
			} else if x, ok := httpRule.GetPattern().(*annotations.HttpRule_Put); ok {
				method = "PUT " + x.Put
			} else if x, ok := httpRule.GetPattern().(*annotations.HttpRule_Post); ok {
				method = "POST " + x.Post
			} else if x, ok := httpRule.GetPattern().(*annotations.HttpRule_Delete); ok {
				method = "DELETE " + x.Delete
			} else if x, ok := httpRule.GetPattern().(*annotations.HttpRule_Patch); ok {
				method = "PATCH " + x.Patch
			}
		}

		doJournal := false
		if methodHandler, ok := getMethodHandler(fullMethod); ok {
			if methodHandler.Journal != nil && *methodHandler.Journal {
				doJournal = true
			}

			if methodHandler.MetricsAlias != nil && *methodHandler.MetricsAlias != "" {
				method = *methodHandler.MetricsAlias
			}
		}

		defer func() { // double recover for safety
			if p := recover(); p != nil {
				errVerbose := fmt.Sprintf("got double panic => error: %+v", errors.Panic(p))
				notify(&adapter.AlertMessage{
					ProjectName:  projectName,
					JournalID:    journalID,
					ErrorVerbose: errVerbose,
					Timestamp:    time.Now(),
				})

				err = status.New(codes.Internal, "got double panic").Err()
				logger.Error(fmt.Sprintf("%s %s", journalID, errVerbose))
			}
		}()

		defer func() {
			if p := recover(); p != nil {
				errVerbose := fmt.Sprintf("got panic => error: %+v", errors.Panic(p))
				notify(&adapter.AlertMessage{
					ProjectName:  projectName,
					JournalID:    journalID,
					ErrorVerbose: errVerbose,
					Timestamp:    time.Now(),
				})

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
									mp[key] = queryUnescape(values[0])
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
					logger.Info("gateway unary interceptor", zap.Any("journal", marshalJournal(journal)))

				} else {
					logger.Error("gateway unary interceptor", zap.Any("journal", marshalJournal(journal)))
				}
			}

			if metrics != nil {
				if err == nil {
					httpRequestSuccessCounter.WithLabelValues(method).Inc()
					httpRequestSuccessDurationHistogram.WithLabelValues(method).Observe(time.Since(ts).Seconds())

				} else {
					s, _ := status.FromError(err)
					code := s.Code().String()

					httpRequestErrorCounter.WithLabelValues(method, code).Inc()
					httpRequestErrorDurationHistogram.WithLabelValues(method, code).Observe(time.Since(ts).Seconds())
				}
			}
		}()

		serviceName := strings.Split(fullMethod, "/")[1]

		var whitelistingValidator adapter.WhitelistingHandler
		if serviceHandler, ok := getServiceHandler(serviceName); ok && serviceHandler.Whitelisting != nil && *serviceHandler.Whitelisting != "" {
			whitelistingValidator, _ = getWhitelistingHandler(*serviceHandler.Whitelisting)
		}
		if methodHandler, ok := getMethodHandler(fullMethod); ok && methodHandler.Whitelisting != nil && *methodHandler.Whitelisting != "" {
			whitelistingValidator, _ = getWhitelistingHandler(*methodHandler.Whitelisting)
		}

		if whitelistingValidator != nil {
			ok, err := whitelistingValidator(meta.Get(XForwardedFor)[0])
			if err != nil {
				notify(&adapter.AlertMessage{
					ProjectName:  projectName,
					JournalID:    journalID,
					ErrorVerbose: fmt.Sprintf("%+v", err),
					Timestamp:    time.Now(),
				})

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
			if s.Code() == codes.Unavailable {
				notify(&adapter.AlertMessage{
					ProjectName:  projectName,
					JournalID:    journalID,
					ErrorVerbose: s.Proto().String(),
					Timestamp:    time.Now(),
				})
			}
		}

		return
	}
}
