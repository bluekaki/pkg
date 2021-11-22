package interceptor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/id"
	"github.com/bluekaki/pkg/pbutil"
	"github.com/bluekaki/pkg/vv/internal/pb"
	"github.com/bluekaki/pkg/vv/proposal"

	protoV1 "github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// UnaryClientInterceptor unary interceptor for client
func UnaryClientInterceptor(logger *zap.Logger, notify proposal.NotifyHandler, signer proposal.Signer, projectName string) grpc.UnaryClientInterceptor {

	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		ts := time.Now()

		doJournal := false
		ignore := false
		if methodHandler, ok := getMethodHandler(method); ok {
			if methodHandler.Journal != nil && *methodHandler.Journal {
				doJournal = true
			}

			if methodHandler.Ignore != nil && *methodHandler.Ignore {
				ignore = true
			}
		}

		var journalID string
		meta, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			journalID = id.JournalID()
			meta = metadata.New(nil)

		} else if values := meta.Get(JournalID); len(values) == 0 || values[0] == "" {
			journalID = id.JournalID()
		} else {
			journalID = values[0]
		}

		meta.Set(JournalID, journalID)
		ctx = metadata.NewOutgoingContext(ctx, meta)

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

			if !ignore {
				journal := &pb.Journal{
					Id: journalID,
					Request: &pb.Request{
						Restapi: false,
						Method:  method,
						Metadata: func() map[string]string {
							if !doJournal {
								return nil
							}

							mp := make(map[string]string)
							for key, values := range meta {
								mp[key] = values[0]
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
							if !doJournal || err != nil || reply == nil {
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

					for _, detail := range s.Details() {
						if stack, ok := detail.(*pb.Stack); ok {
							journal.Response.ErrorVerbose = stack.Verbose
						}
					}

					err = status.New(s.Code(), s.Message()).Err() // reset detail
				}

				journal.CostSeconds = time.Since(ts).Seconds()

				if err == nil {
					logger.Info("client unary interceptor", zap.Any("journal", marshalJournal(journal)))

				} else {
					logger.Error("client unary interceptor", zap.Any("journal", marshalJournal(journal)))
				}
			}
		}()

		if signer != nil {
			var raw json.RawMessage
			if req != nil {
				if raw, err = pbutil.ProtoMessage2JSON(req.(protoV1.Message)); err != nil {
					return
				}
			}

			var signature, date string
			if signature, date, err = signer(method, []byte(raw)); err != nil {
				return
			}

			meta.Set(AuthorizationProxy, signature)
			meta.Set(Date, date)
			ctx = metadata.NewOutgoingContext(ctx, meta)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor stream interceptor for client
func StreamClientInterceptor(logger *zap.Logger, notify proposal.NotifyHandler, signer proposal.Signer, projectName string) grpc.StreamClientInterceptor {

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, fullMethod string, streamer grpc.Streamer, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
		ts := time.Now()

		doJournal := false
		ignore := false
		if methodHandler, ok := getMethodHandler(fullMethod); ok {
			if methodHandler.Journal != nil && *methodHandler.Journal {
				doJournal = true
			}

			if methodHandler.Ignore != nil && *methodHandler.Ignore {
				ignore = true
			}
		}

		var journalID string
		meta, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			journalID = id.JournalID()
			meta = metadata.New(nil)

		} else if values := meta.Get(JournalID); len(values) == 0 || values[0] == "" {
			journalID = id.JournalID()

		} else {
			journalID = values[0]
		}

		meta.Set(JournalID, journalID)
		ctx = metadata.NewOutgoingContext(ctx, meta)

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

			if !ignore {
				journal := &pb.Journal{
					Id: journalID,
					Label: &pb.Lable{
						Desc: "Stream",
					},
					Request: &pb.Request{
						Restapi: false,
						Method:  fullMethod,
						Metadata: func() map[string]string {
							if !doJournal {
								return nil
							}

							mp := make(map[string]string)
							for key, values := range meta {
								mp[key] = values[0]
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

					err = status.New(s.Code(), s.Message()).Err() // reset detail
				}

				journal.CostSeconds = time.Since(ts).Seconds()

				if err == nil {
					logger.Info("client stream interceptor", zap.Any("journal", marshalJournal(journal)))

				} else {
					logger.Error("client stream interceptor", zap.Any("journal", marshalJournal(journal)))
				}
			}
		}()

		if signer != nil {
			var signature, date string
			if signature, date, err = signer(fullMethod, []byte(journalID)); err != nil {
				return
			}

			meta.Set(AuthorizationProxy, signature)
			meta.Set(Date, date)
			ctx = metadata.NewOutgoingContext(ctx, meta)
		}

		s, err := streamer(ctx, desc, cc, fullMethod, opts...)
		if err != nil {
			return nil, err
		}

		return &streamClientInterceptor{
			ClientStream: s,
			logger:       logger,

			journalID: journalID,

			ignore:    ignore,
			doJournal: doJournal,
			method:    fullMethod,
		}, nil
	}
}

type streamClientInterceptor struct {
	grpc.ClientStream
	logger *zap.Logger

	journalID string

	counter struct {
		send uint32
		recv uint32
	}

	ignore    bool
	doJournal bool
	method    string
}

func (s *streamClientInterceptor) SendMsg(m interface{}) (err error) {
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
				Request: &pb.Request{
					Method: s.method,
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

			s.logger.Info("client stream/send interceptor", zap.Any("journal", marshalJournal(journal)))
		}
	}()

	return s.ClientStream.SendMsg(m)
}

func (s *streamClientInterceptor) RecvMsg(m interface{}) (err error) {
	ts := time.Now()
	defer func() {
		if err == io.EOF {
			return
		}

		s.counter.recv++

		if !s.ignore {
			journal := &pb.Journal{
				Id: s.journalID,
				Label: &pb.Lable{
					Sequence: s.counter.recv,
					Desc:     "RecvMsg",
				},
				Response: &pb.Response{
					Code: codes.OK.String(),
					Payload: func() *anypb.Any {
						if !s.doJournal || err != nil || m == nil {
							return nil
						}

						any, _ := anypb.New(m.(proto.Message))
						return any
					}(),
				},
				Success:     err == nil,
				CostSeconds: time.Since(ts).Seconds(),
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

				err = status.New(s.Code(), s.Message()).Err() // reset detail
			}

			s.logger.Info("client stream/recv interceptor", zap.Any("journal", marshalJournal(journal)))
		}
	}()

	return s.ClientStream.RecvMsg(m)
}
