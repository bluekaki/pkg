package interceptor

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/bluekaki/pkg/pbutil"
	"github.com/bluekaki/pkg/vv/internal/protos/gen"

	protoV1 "github.com/golang/protobuf/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Sign signs the message
type Sign func(fullMethod string, message []byte) (auth, date string, err error)

// NewClientInterceptor create a client interceptor
func NewClientInterceptor(sign Sign, logger *zap.Logger, marshalJournal bool, notify notifyHandler) *ClientInterceptor {
	return &ClientInterceptor{
		sign:           sign,
		logger:         logger,
		marshalJournal: marshalJournal,
		notify:         notify,
	}
}

// ClientInterceptor the client's interceptor
type ClientInterceptor struct {
	sign           Sign
	logger         *zap.Logger
	marshalJournal bool
	notify         notifyHandler
}

// UnaryInterceptor a interceptor for client unary operations
func (c *ClientInterceptor) UnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	ts := time.Now()

	var journalID string
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		journalID = GenJournalID()
	} else if values := meta.Get(JournalID); len(values) == 0 || values[0] == "" {
		journalID = GenJournalID()
	} else {
		journalID = values[0]
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	md := make(metadata.MD)
	md.Set(JournalID, journalID)

	ctx = metadata.NewOutgoingContext(ctx, md)

	defer func() { // double recover for safety
		if p := recover(); p != nil {
			err = status.New(codes.Internal, "got double panic").Err()
			c.logger.Error(fmt.Sprintf("got double panic => error: %+v\n%s", p, string(debug.Stack())))
		}
	}()

	defer func() {
		if p := recover(); p != nil {
			msg := fmt.Sprintf("got panic => error: %+v", p)
			info := string(debug.Stack())
			if c.notify != nil {
				c.notify("got panic", msg, info, journalID)
			}

			s, _ := status.New(codes.Internal, msg).WithDetails(&pb.Stack{Info: info})
			err = s.Err()
		}

		journal := &pb.Journal{
			Id: journalID,
			Request: &pb.Request{
				Restapi: false,
				Method:  method,
				Metadata: func() map[string]string {
					mp := make(map[string]string)
					for key, values := range md {
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

			journal.Response.Details = make([]*anypb.Any, len(s.Details()))
			for i, detail := range s.Details() {
				journal.Response.Details[i], _ = anypb.New(detail.(proto.Message))
			}

			err = status.New(s.Code(), s.Message()).Err() // reset detail
		}

		journal.CostSeconds = time.Since(ts).Seconds()

		var json interface{}
		if c.marshalJournal {
			json, _ = pbutil.ProtoMessage2JSON(journal)
		} else {
			json, _ = pbutil.ProtoMessage2Map(journal)
		}

		if err == nil {
			c.logger.Info("client unary interceptor", zap.Any("journal", json))
		} else {
			c.logger.Error("client unary interceptor", zap.Any("journal", json))
		}
	}()

	if c.sign != nil {
		var raw string
		if req != nil {
			if raw, err = pbutil.ProtoMessage2JSON(req.(protoV1.Message)); err != nil {
				return
			}
		}

		var signature, date string
		if signature, date, err = c.sign(method, []byte(raw)); err != nil {
			return
		}

		md.Set(Date, date)
		md.Set(ProxyAuthorization, signature)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}

type clientWrappedStream struct {
	grpc.ClientStream
}

func (c *clientWrappedStream) RecvMsg(m interface{}) error {
	return c.ClientStream.RecvMsg(m)
}

func (c *clientWrappedStream) SendMsg(m interface{}) error {
	return c.ClientStream.SendMsg(m)
}

// StreamInterceptor a interceptor for client stream operations
func (c *ClientInterceptor) StreamInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (stream grpc.ClientStream, err error) {
	defer func() {
		if p := recover(); p != nil {
			s, _ := status.New(codes.Internal, fmt.Sprintf("%+v", p)).WithDetails(&pb.Stack{Info: string(debug.Stack())})
			err = s.Err()
		}
	}()

	stream, err = streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return
	}

	return &clientWrappedStream{ClientStream: stream}, nil
}
