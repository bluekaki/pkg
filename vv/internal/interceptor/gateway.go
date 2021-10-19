package interceptor

import (
	"context"

	"github.com/bluekaki/pkg/vv/pkg/adapter"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func UnaryClientInterceptor(logger *zap.Logger, notify adapter.NotifyHandler) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {

		return nil
	}
}
