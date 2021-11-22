package interceptor

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/id"
	"github.com/bluekaki/pkg/pbutil"
	"github.com/bluekaki/pkg/vv/internal/pb"
	"github.com/bluekaki/pkg/vv/internal/pkg/multipart"
	"github.com/bluekaki/pkg/vv/proposal"

	protoV1 "github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// SessionUserinfo mark userinfo in context
type SessionUserinfo struct{}

// SignatureIdentifier mark identifier in context
type SignatureIdentifier struct{}

var _ proposal.Payload = (*restPayload)(nil)
var _ proposal.Payload = (*grpcPayload)(nil)

type restPayload struct {
	journalID string
	service   string
	date      string
	method    string
	uri       string
	body      []byte
}

func (r *restPayload) JournalID() string {
	return r.journalID
}

func (r *restPayload) ForwardedByGrpcGateway() bool {
	return true
}

func (r *restPayload) Service() string {
	return r.service
}

func (r *restPayload) Date() string {
	return r.date
}

func (r *restPayload) Method() string {
	return r.method
}

func (r *restPayload) URI() string {
	return r.uri
}

func (r *restPayload) Body() []byte {
	return r.body
}

type grpcPayload struct {
	journalID string
	service   string
	date      string
	method    string
	uri       string
	body      []byte
}

func (g *grpcPayload) JournalID() string {
	return g.journalID
}

func (g *grpcPayload) ForwardedByGrpcGateway() bool {
	return false
}

func (g *grpcPayload) Service() string {
	return g.service
}

func (g *grpcPayload) Date() string {
	return g.date
}

func (g *grpcPayload) Method() string {
	return g.method
}

func (g *grpcPayload) URI() string {
	return g.uri
}

func (g *grpcPayload) Body() []byte {
	return g.body
}

var serverInitMetricsOnce sync.Once

// UnaryServerInterceptor unary interceptor for server
func UnaryServerInterceptor(logger *zap.Logger, notify proposal.NotifyHandler, metrics func(http.Handler), projectName string) grpc.UnaryServerInterceptor {
	if metrics != nil {
		serverInitMetricsOnce.Do(func() {
			metrics(promhttp.Handler())
		})
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		ts := time.Now()

		fullMethod := strings.Split(info.FullMethod, "/")
		serviceName := fullMethod[1]

		doJournal := false
		ignore := false
		method := info.FullMethod
		if methodHandler, ok := getMethodHandler(info.FullMethod); ok {
			if methodHandler.Journal != nil && *methodHandler.Journal {
				doJournal = true
			}

			if methodHandler.Ignore != nil && *methodHandler.Ignore {
				ignore = true
			}

			if methodHandler.MetricsAlias != nil && *methodHandler.MetricsAlias != "" {
				method = *methodHandler.MetricsAlias
			}
		}

		var journalID string
		meta, _ := metadata.FromIncomingContext(ctx)
		if values := meta.Get(JournalID); len(values) == 0 || values[0] == "" {
			journalID = id.JournalID()
			meta.Set(JournalID, journalID)
			ctx = metadata.NewOutgoingContext(ctx, meta)

		} else {
			journalID = values[0]
		}

		defer func() {
			grpc.SetHeader(ctx, metadata.Pairs(runtime.MetadataHeaderPrefix+JournalID, journalID))
			grpc.SetHeader(ctx, metadata.Pairs(JournalID, journalID))

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

			var statusCode *pb.Code
			if err != nil {
				switch err.(type) {
				case proposal.BzError:
					bzErr := err.(proposal.BzError)
					statusCode = &pb.Code{HttpStatus: uint32(bzErr.HTTPCode())}
					s, _ := status.New(codes.Code(bzErr.BzCode()), bzErr.Desc()).WithDetails(&pb.Stack{Verbose: fmt.Sprintf("%+v", bzErr.StackErr())})
					err = s.Err()

				case proposal.AlertError:
					alertErr := err.(proposal.AlertError)

					alert := alertErr.AlertMessage()
					alert.ProjectName = projectName
					alert.JournalID = journalID
					notify(alert)

					bzErr := alertErr.BzError()
					statusCode = &pb.Code{HttpStatus: uint32(bzErr.HTTPCode())}
					s, _ := status.New(codes.Code(bzErr.BzCode()), bzErr.Desc()).WithDetails(&pb.Stack{Verbose: fmt.Sprintf("%+v", bzErr.StackErr())})
					err = s.Err()
				}
			}

			if !ignore {
				journal := &pb.Journal{
					Id: journalID,
					Request: &pb.Request{
						Restapi: forwardedByGrpcGateway(meta),
						Method:  info.FullMethod,
						Metadata: func() map[string]string {
							if !doJournal {
								return nil
							}

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
							if !doJournal || req == nil {
								return nil
							}

							any, _ := anypb.New(req.(proto.Message))
							return any
						}(),
					},
					Response: &pb.Response{
						Code: codes.OK.String(),
						Payload: func() *anypb.Any {
							if !doJournal || err != nil || resp == nil {
								return nil
							}

							any, _ := anypb.New(resp.(proto.Message))
							return any
						}(),
					},
					Success: err == nil,
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

					s = status.New(s.Code(), s.Message()) // reset detail
					if statusCode != nil {
						s, _ = s.WithDetails(statusCode)
					}

					err = s.Err()
				}

				journal.CostSeconds = time.Since(ts).Seconds()

				if err == nil {
					logger.Info("server unary interceptor", zap.Any("journal", marshalJournal(journal)))

				} else {
					logger.Error("server unary interceptor", zap.Any("journal", marshalJournal(journal)))
				}
			}

			if metrics != nil {
				if err == nil {
					grpcRequestSuccessCounter.WithLabelValues(method).Inc()
					grpcRequestSuccessDurationHistogram.WithLabelValues(method).Observe(time.Since(ts).Seconds())

				} else {
					s, _ := status.FromError(err)
					code := s.Code().String()

					grpcRequestErrorCounter.WithLabelValues(method, code).Inc()
					grpcRequestErrorDurationHistogram.WithLabelValues(method, code).Observe(time.Since(ts).Seconds())
				}
			}
		}()

		if req != nil {
			if validator, ok := req.(proposal.Validator); ok {
				if err := validator.Validate(); err != nil {
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
			}
		}

		var (
			authorizationValidator      proposal.UserinfoHandler
			authorizationProxyValidator proposal.SignatureHandler
		)

		if serviceHandler, ok := getServiceHandler(serviceName); ok {
			if serviceHandler.Authorization != nil && *serviceHandler.Authorization != "" {
				authorizationValidator, _ = getAuthorizationHandler(*serviceHandler.Authorization)
			}

			if serviceHandler.AuthorizationProxy != nil && *serviceHandler.AuthorizationProxy != "" {
				authorizationProxyValidator, _ = getAuthorizationProxyHandler(*serviceHandler.AuthorizationProxy)
			}
		}

		if methodHandler, ok := getMethodHandler(info.FullMethod); ok {
			if methodHandler.Authorization != nil && *methodHandler.Authorization != "" {
				authorizationValidator, _ = getAuthorizationHandler(*methodHandler.Authorization)
			}

			if methodHandler.AuthorizationProxy != nil && *methodHandler.AuthorizationProxy != "" {
				authorizationProxyValidator, _ = getAuthorizationProxyHandler(*methodHandler.AuthorizationProxy)
			}
		}

		if authorizationValidator == nil && authorizationProxyValidator == nil {
			return handler(ctx, req)
		}

		var auth, authProxy string
		if authHeader := meta.Get(Authorization); len(authHeader) != 0 {
			auth = authHeader[0]
		}
		if authProxyHeader := meta.Get(AuthorizationProxy); len(authProxyHeader) != 0 {
			authProxy = authProxyHeader[0]
		}

		var payload proposal.Payload
		if forwardedByGrpcGateway(meta) {
			payload = &restPayload{
				journalID: journalID,
				service:   serviceName,
				date:      meta.Get(Date)[0],
				method:    meta.Get(Method)[0],
				uri:       meta.Get(URI)[0],
				body: func() []byte {
					if meta.Get(OctetStream)[0] != "" {
						raw, _ := base64.StdEncoding.DecodeString(meta.Get(Body)[0])
						return bytes.Join(multipart.ParseFormData(raw), nil)

					} else {
						return []byte(meta.Get(Body)[0])
					}
				}(),
			}

		} else {
			payload = &grpcPayload{
				journalID: journalID,
				service:   serviceName,
				date: func() string {
					if date := meta.Get(Date); len(date) > 0 {
						return date[0]
					}
					return ""
				}(),
				method: "GRPC",
				uri:    info.FullMethod,
				body: func() []byte {
					if req == nil {
						return nil
					}

					raw, _ := pbutil.ProtoMessage2JSON(req.(protoV1.Message))
					return []byte(raw)
				}(),
			}
		}

		if authorizationValidator != nil {
			userinfo, err := authorizationValidator(auth, payload)
			if err != nil {
				s := status.New(codes.Unauthenticated, codes.Unauthenticated.String())
				s, _ = s.WithDetails(&pb.Stack{Verbose: fmt.Sprintf("%+v", err)})
				return nil, s.Err()
			}
			ctx = context.WithValue(ctx, SessionUserinfo{}, userinfo)
		}

		if authorizationProxyValidator != nil {
			identifier, ok, err := authorizationProxyValidator(authProxy, payload)
			if err != nil {
				s := status.New(codes.PermissionDenied, codes.PermissionDenied.String())
				s, _ = s.WithDetails(&pb.Stack{Verbose: fmt.Sprintf("%+v", err)})
				return nil, s.Err()
			}
			if !ok {
				return nil, status.Error(codes.PermissionDenied, "signature does not match")
			}
			ctx = context.WithValue(ctx, SignatureIdentifier{}, identifier)
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor stream interceptor for server
func StreamServerInterceptor(logger *zap.Logger, notify proposal.NotifyHandler, metrics func(http.Handler), projectName string) grpc.StreamServerInterceptor {
	if metrics != nil {
		serverInitMetricsOnce.Do(func() {
			metrics(promhttp.Handler())
		})
	}

	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		ts := time.Now()

		fullMethod := strings.Split(info.FullMethod, "/")
		serviceName := fullMethod[1]

		doJournal := false
		ignore := false
		method := info.FullMethod
		if methodHandler, ok := getMethodHandler(info.FullMethod); ok {
			if methodHandler.Journal != nil && *methodHandler.Journal {
				doJournal = true
			}

			if methodHandler.Ignore != nil && *methodHandler.Ignore {
				ignore = true
			}

			if methodHandler.MetricsAlias != nil && *methodHandler.MetricsAlias != "" {
				method = *methodHandler.MetricsAlias
			}
		}

		var journalID string
		meta, _ := metadata.FromIncomingContext(stream.Context())
		if values := meta.Get(JournalID); len(values) == 0 || values[0] == "" {
			journalID = id.JournalID()

		} else {
			journalID = values[0]
		}

		stream.SendHeader(metadata.Pairs(JournalID, journalID))
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

			var statusCode *pb.Code
			if err != nil {
				switch err.(type) {
				case proposal.BzError:
					bzErr := err.(proposal.BzError)
					statusCode = &pb.Code{HttpStatus: uint32(bzErr.HTTPCode())}
					s, _ := status.New(codes.Code(bzErr.BzCode()), bzErr.Desc()).WithDetails(&pb.Stack{Verbose: fmt.Sprintf("%+v", bzErr.StackErr())})
					err = s.Err()

				case proposal.AlertError:
					alertErr := err.(proposal.AlertError)

					alert := alertErr.AlertMessage()
					alert.ProjectName = projectName
					alert.JournalID = journalID
					notify(alert)

					bzErr := alertErr.BzError()
					statusCode = &pb.Code{HttpStatus: uint32(bzErr.HTTPCode())}
					s, _ := status.New(codes.Code(bzErr.BzCode()), bzErr.Desc()).WithDetails(&pb.Stack{Verbose: fmt.Sprintf("%+v", bzErr.StackErr())})
					err = s.Err()
				}
			}

			if !ignore {
				journal := &pb.Journal{
					Id: journalID,
					Label: &pb.Lable{
						Desc: "Stream",
					},
					Request: &pb.Request{
						Restapi: forwardedByGrpcGateway(meta),
						Method:  info.FullMethod,
						Metadata: func() map[string]string {
							if !doJournal {
								return nil
							}

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

					for _, detail := range s.Details() {
						if stack, ok := detail.(*pb.Stack); ok {
							journal.Response.ErrorVerbose = stack.Verbose
						}
					}

					s = status.New(s.Code(), s.Message()) // reset detail
					if statusCode != nil {
						s, _ = s.WithDetails(statusCode)
					}

					err = s.Err()
				}

				journal.CostSeconds = time.Since(ts).Seconds()

				if err == nil {
					logger.Info("server stream interceptor", zap.Any("journal", marshalJournal(journal)))

				} else {
					logger.Error("server stream interceptor", zap.Any("journal", marshalJournal(journal)))
				}
			}

			if metrics != nil {
				if err == nil {
					grpcRequestSuccessCounter.WithLabelValues(method).Inc()
					grpcRequestSuccessDurationHistogram.WithLabelValues(method).Observe(time.Since(ts).Seconds())

				} else {
					s, _ := status.FromError(err)
					code := s.Code().String()

					grpcRequestErrorCounter.WithLabelValues(method, code).Inc()
					grpcRequestErrorDurationHistogram.WithLabelValues(method, code).Observe(time.Since(ts).Seconds())
				}
			}
		}()

		var (
			authorizationValidator      proposal.UserinfoHandler
			authorizationProxyValidator proposal.SignatureHandler
		)

		if serviceHandler, ok := getServiceHandler(serviceName); ok {
			if serviceHandler.Authorization != nil && *serviceHandler.Authorization != "" {
				authorizationValidator, _ = getAuthorizationHandler(*serviceHandler.Authorization)
			}

			if serviceHandler.AuthorizationProxy != nil && *serviceHandler.AuthorizationProxy != "" {
				authorizationProxyValidator, _ = getAuthorizationProxyHandler(*serviceHandler.AuthorizationProxy)
			}
		}

		if methodHandler, ok := getMethodHandler(info.FullMethod); ok {
			if methodHandler.Authorization != nil && *methodHandler.Authorization != "" {
				authorizationValidator, _ = getAuthorizationHandler(*methodHandler.Authorization)
			}

			if methodHandler.AuthorizationProxy != nil && *methodHandler.AuthorizationProxy != "" {
				authorizationProxyValidator, _ = getAuthorizationProxyHandler(*methodHandler.AuthorizationProxy)
			}
		}

		if authorizationValidator == nil && authorizationProxyValidator == nil {
			return handler(srv, &streamServerInterceptor{
				ServerStream: stream,
				logger:       logger,

				journalID: journalID,

				ignore:    ignore,
				doJournal: doJournal,
				restapi:   forwardedByGrpcGateway(meta),
				method:    info.FullMethod,
			})
		}

		var auth, authProxy string
		if authHeader := meta.Get(Authorization); len(authHeader) != 0 {
			auth = authHeader[0]
		}
		if authProxyHeader := meta.Get(AuthorizationProxy); len(authProxyHeader) != 0 {
			authProxy = authProxyHeader[0]
		}

		var payload proposal.Payload
		if forwardedByGrpcGateway(meta) {
			payload = &restPayload{
				journalID: journalID,
				service:   serviceName,
				date:      meta.Get(Date)[0],
				method:    meta.Get(Method)[0],
				uri:       meta.Get(URI)[0],
				body: func() []byte {
					if meta.Get(OctetStream)[0] != "" {
						raw, _ := base64.StdEncoding.DecodeString(meta.Get(Body)[0])
						return bytes.Join(multipart.ParseFormData(raw), nil)

					} else {
						return []byte(meta.Get(Body)[0])
					}
				}(),
			}

		} else {
			payload = &grpcPayload{
				journalID: journalID,
				service:   serviceName,
				date: func() string {
					if date := meta.Get(Date); len(date) > 0 {
						return date[0]
					}
					return ""
				}(),
				method: "GRPC",
				uri:    info.FullMethod,
				body: func() []byte {
					return []byte(journalID)
				}(),
			}
		}

		var userinfo interface{}
		if authorizationValidator != nil {
			userinfo, err = authorizationValidator(auth, payload)
			if err != nil {
				s := status.New(codes.Unauthenticated, codes.Unauthenticated.String())
				s, _ = s.WithDetails(&pb.Stack{Verbose: fmt.Sprintf("%+v", err)})
				return s.Err()
			}
		}

		var identifier string
		if authorizationProxyValidator != nil {
			var ok bool
			identifier, ok, err = authorizationProxyValidator(authProxy, payload)
			if err != nil {
				s := status.New(codes.PermissionDenied, codes.PermissionDenied.String())
				s, _ = s.WithDetails(&pb.Stack{Verbose: fmt.Sprintf("%+v", err)})
				return s.Err()
			}
			if !ok {
				return status.Error(codes.PermissionDenied, "signature does not match")
			}
		}

		return handler(srv, &streamServerInterceptor{
			ServerStream: stream,
			logger:       logger,

			journalID:  journalID,
			userinfo:   userinfo,
			identifier: identifier,

			ignore:    ignore,
			doJournal: doJournal,
			restapi:   forwardedByGrpcGateway(meta),
			method:    info.FullMethod,
		})
	}
}

type streamServerInterceptor struct {
	grpc.ServerStream
	logger *zap.Logger

	journalID  string
	userinfo   interface{}
	identifier string

	counter struct {
		send uint32
		recv uint32
	}

	ignore    bool
	doJournal bool
	restapi   bool
	method    string
}

func (s *streamServerInterceptor) Context() context.Context {
	meta, _ := metadata.FromIncomingContext(s.ServerStream.Context())
	meta.Set(JournalID, s.journalID)

	ctx := metadata.NewIncomingContext(s.ServerStream.Context(), meta)
	ctx = context.WithValue(ctx, SessionUserinfo{}, s.userinfo)
	ctx = context.WithValue(ctx, SignatureIdentifier{}, s.identifier)

	return ctx
}

func (s *streamServerInterceptor) SendMsg(m interface{}) (err error) {
	s.counter.send++

	ts := time.Now()
	defer func() {
		if !s.ignore {
			journal := &pb.Journal{
				Id: s.journalID,
				Label: &pb.Lable{
					Sequence: s.counter.send,
					Desc:     "SendMsg",
				},
				Response: &pb.Response{
					Payload: func() *anypb.Any {
						if !s.doJournal || m == nil {
							return nil
						}

						any, _ := anypb.New(m.(proto.Message))
						return any
					}(),
				},
				Success:     err == nil,
				CostSeconds: time.Since(ts).Seconds(),
			}

			s.logger.Info("server stream/send interceptor", zap.Any("journal", marshalJournal(journal)))
		}
	}()

	return s.ServerStream.SendMsg(m)
}

func (s *streamServerInterceptor) RecvMsg(m interface{}) (err error) {
	ts := time.Now()
	defer func() {
		if err == io.EOF {
			return
		}

		if err == nil && m != nil {
			if validator, ok := m.(proposal.Validator); ok {
				if err = validator.Validate(); err != nil {
					err = status.Error(codes.InvalidArgument, err.Error())
				}
			}
		}

		s.counter.recv++

		if !s.ignore {
			journal := &pb.Journal{
				Id: s.journalID,
				Label: &pb.Lable{
					Sequence: s.counter.recv,
					Desc:     "RecvMsg",
				},
				Request: &pb.Request{
					Restapi: s.restapi,
					Method:  s.method,
					Payload: func() *anypb.Any {
						if !s.doJournal || m == nil {
							return nil
						}

						any, _ := anypb.New(m.(proto.Message))
						return any
					}(),
				},
				Success:     err == nil,
				CostSeconds: time.Since(ts).Seconds(),
			}

			s.logger.Info("server stream/recv interceptor", zap.Any("journal", marshalJournal(journal)))
		}
	}()

	return s.ServerStream.RecvMsg(m)
}
