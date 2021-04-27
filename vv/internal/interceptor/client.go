package interceptor

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/byepichi/pkg/pbutil"
	"github.com/byepichi/pkg/vv/internal/protos/gen"

	protoV1 "github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Sign signs the message
type Sign func(fullMethod string, message []byte) (auth, date string, err error)

// NewClientInterceptor create a client interceptor
func NewClientInterceptor(sign Sign) *ClientInterceptor {
	return &ClientInterceptor{sign: sign}
}

// ClientInterceptor the client's interceptor
type ClientInterceptor struct {
	sign Sign
}

// UnaryInterceptor a interceptor for client unary operations
func (c *ClientInterceptor) UnaryInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	defer func() {
		if p := recover(); p != nil {
			s, _ := status.New(codes.Internal, fmt.Sprintf("%+v", p)).WithDetails(&pb.Stack{Info: string(debug.Stack())})
			err = s.Err()
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

		meta, _ := metadata.FromOutgoingContext(ctx)
		if meta == nil {
			meta = make(metadata.MD)
		}

		meta.Set(Date, date)
		meta.Set(ProxyAuthorization, signature)
		ctx = metadata.NewOutgoingContext(ctx, meta)
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
