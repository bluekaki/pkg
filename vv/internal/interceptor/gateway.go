package interceptor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/vv/internal/pb"
	"github.com/bluekaki/pkg/vv/proposal"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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

var gatewayInitMetricsOnce sync.Once

// UnaryGatewayInterceptor unary interceptor for gateway
func UnaryGatewayInterceptor(logger *zap.Logger, notify proposal.NotifyHandler, metrics func(http.Handler), projectName string) grpc.UnaryClientInterceptor {
	if metrics != nil {
		gatewayInitMetricsOnce.Do(func() {
			metrics(promhttp.Handler())
		})
	}

	return func(ctx context.Context, fullMethod string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		ts := time.Now()

		meta, _ := metadata.FromOutgoingContext(ctx)
		meta.Set(gwHeader.key, gwHeader.value)

		journalID := meta.Get(JournalID)[0]

		method := fullMethod
		if httpRule, ok := getHTTPRule(fullMethod); ok {
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

		defer func() {
			if p := recover(); p != nil {
				errVerbose := fmt.Sprintf("got panic => error: %+v", errors.Panic(p))
				notify(&proposal.AlertMessage{
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

					var customStatus *runtime.HTTPStatusError
					for _, detail := range s.Details() {
						switch detail.(type) {
						case *pb.Stack:
							journal.Response.ErrorVerbose = detail.(*pb.Stack).Verbose

						case *pb.Code:
							customStatus = &runtime.HTTPStatusError{HTTPStatus: int(detail.(*pb.Code).HttpStatus)}
						}
					}

					err = status.New(s.Code(), s.Message()).Err() // reset detail
					if customStatus != nil {
						customStatus.Err = err
						err = customStatus
					}
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

		var whitelistingValidator proposal.WhitelistingHandler
		if serviceHandler, ok := getServiceHandler(serviceName); ok && serviceHandler.Whitelisting != nil && *serviceHandler.Whitelisting != "" {
			whitelistingValidator, _ = getWhitelistingHandler(*serviceHandler.Whitelisting)
		}
		if methodHandler, ok := getMethodHandler(fullMethod); ok && methodHandler.Whitelisting != nil && *methodHandler.Whitelisting != "" {
			whitelistingValidator, _ = getWhitelistingHandler(*methodHandler.Whitelisting)
		}

		if whitelistingValidator != nil {
			ok, err := whitelistingValidator(meta.Get(XForwardedFor)[0])
			if err != nil {
				errorVerbose := fmt.Sprintf("%+v", err)
				notify(&proposal.AlertMessage{
					ProjectName:  projectName,
					JournalID:    journalID,
					ErrorVerbose: errorVerbose,
					Timestamp:    time.Now(),
				})

				s := status.New(codes.Aborted, codes.Aborted.String())
				s, _ = s.WithDetails(&pb.Stack{Verbose: errorVerbose})
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
				notify(&proposal.AlertMessage{
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

// StreamGatewayInterceptor stream interceptor for gateway
func StreamGatewayInterceptor(logger *zap.Logger, notify proposal.NotifyHandler, metrics func(http.Handler), projectName string) grpc.StreamClientInterceptor {
	if metrics != nil {
		gatewayInitMetricsOnce.Do(func() {
			metrics(promhttp.Handler())
		})
	}

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, fullMethod string, streamer grpc.Streamer, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
		ts := time.Now()

		meta, _ := metadata.FromOutgoingContext(ctx)
		meta.Set(gwHeader.key, gwHeader.value)

		journalID := meta.Get(JournalID)[0]

		method := fullMethod
		if httpRule, ok := getHTTPRule(fullMethod); ok {
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

		defer func() {
			if p := recover(); p != nil {
				errVerbose := fmt.Sprintf("got panic => error: %+v", errors.Panic(p))
				notify(&proposal.AlertMessage{
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
					Label: &pb.Lable{
						Desc: "Stream",
					},
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
					},
					Response: &pb.Response{
						Code: codes.OK.String(),
					},
					Success: err == nil,
				}

				if err != nil {
					s, _ := status.FromError(err)
					journal.Response.Code = s.Code().String()
					journal.Response.Message = s.Message()

					var customStatus *runtime.HTTPStatusError
					for _, detail := range s.Details() {
						switch detail.(type) {
						case *pb.Stack:
							journal.Response.ErrorVerbose = detail.(*pb.Stack).Verbose

						case *pb.Code:
							customStatus = &runtime.HTTPStatusError{HTTPStatus: int(detail.(*pb.Code).HttpStatus)}
						}
					}

					err = status.New(s.Code(), s.Message()).Err() // reset detail
					if customStatus != nil {
						customStatus.Err = err
						err = customStatus
					}
				}

				journal.CostSeconds = time.Since(ts).Seconds()

				if err == nil {
					logger.Info("gateway stream interceptor", zap.Any("journal", marshalJournal(journal)))

				} else {
					logger.Error("gateway stream interceptor", zap.Any("journal", marshalJournal(journal)))
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

		var whitelistingValidator proposal.WhitelistingHandler
		if serviceHandler, ok := getServiceHandler(serviceName); ok && serviceHandler.Whitelisting != nil && *serviceHandler.Whitelisting != "" {
			whitelistingValidator, _ = getWhitelistingHandler(*serviceHandler.Whitelisting)
		}
		if methodHandler, ok := getMethodHandler(fullMethod); ok && methodHandler.Whitelisting != nil && *methodHandler.Whitelisting != "" {
			whitelistingValidator, _ = getWhitelistingHandler(*methodHandler.Whitelisting)
		}

		if whitelistingValidator != nil {
			ok, err := whitelistingValidator(meta.Get(XForwardedFor)[0])
			if err != nil {
				errorVerbose := fmt.Sprintf("%+v", err)
				notify(&proposal.AlertMessage{
					ProjectName:  projectName,
					JournalID:    journalID,
					ErrorVerbose: errorVerbose,
					Timestamp:    time.Now(),
				})

				s := status.New(codes.Aborted, codes.Aborted.String())
				s, _ = s.WithDetails(&pb.Stack{Verbose: errorVerbose})
				return nil, s.Err()
			}

			if !ok {
				return nil, status.Error(codes.Aborted, "ip does not allow access")
			}
		}

		s, err := streamer(metadata.NewOutgoingContext(ctx, meta), desc, cc, fullMethod, opts...)
		if err != nil {
			return nil, err
		}

		return &streamGatewayInterceptor{
			ClientStream: s,
			logger:       logger,
			doJournal:    doJournal,
			journalID:    journalID,
		}, nil
	}
}

type streamGatewayInterceptor struct {
	grpc.ClientStream
	logger    *zap.Logger
	doJournal bool
	journalID string
	counter   struct {
		send uint32
		recv uint32
	}
}

func (s *streamGatewayInterceptor) SendMsg(m interface{}) (err error) {
	s.counter.recv++

	if s.doJournal {
		journal := &pb.Journal{
			Id: s.journalID,
			Label: &pb.Lable{
				Sequence: s.counter.recv,
				Desc:     "RecvMsg",
			},
			Request: &pb.Request{
				Payload: func() *anypb.Any {
					if m == nil {
						return nil
					}

					any, _ := anypb.New(m.(proto.Message))
					return any
				}(),
			},
		}

		s.logger.Info("gateway stream/recv interceptor", zap.Any("journal", marshalJournal(journal)))
	}

	return s.ClientStream.SendMsg(m)
}

func (s *streamGatewayInterceptor) RecvMsg(m interface{}) (err error) {
	defer func() {
		if err == io.EOF {
			return
		}

		s.counter.send++

		if s.doJournal {
			journal := &pb.Journal{
				Id: s.journalID,
				Label: &pb.Lable{
					Sequence: s.counter.send,
					Desc:     "SendMsg",
				},
				Response: &pb.Response{
					Code: codes.OK.String(),
					Payload: func() *anypb.Any {
						if m == nil {
							return nil
						}

						any, _ := anypb.New(m.(proto.Message))
						return any
					}(),
				},
			}

			if err != nil {
				s, _ := status.FromError(err)
				journal.Response.Code = s.Code().String()
				journal.Response.Message = s.Message()

				for _, detail := range s.Details() {
					if stack, ok := detail.(*pb.Stack); ok {
						journal.Response.ErrorVerbose = stack.Verbose
					}
				}
			}

			s.logger.Info("gateway stream/send interceptor", zap.Any("journal", marshalJournal(journal)))
		}
	}()

	return s.ClientStream.RecvMsg(m)
}
