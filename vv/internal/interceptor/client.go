package interceptor

import (
	"context"
	"encoding/json"
	"fmt"
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

			journal := &pb.Journal{
				Id: journalID,
				Request: &pb.Request{
					Restapi: false,
					Method:  method,
					Metadata: func() map[string]string {
						mp := make(map[string]string)
						for key, values := range meta {
							mp[key] = values[0]
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
func StreamClientInterceptor() grpc.StreamClientInterceptor {

	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
		return nil, errors.New("not currently supported")
	}
}
