package vv

import (
	"context"
	stderr "errors"
	"fmt"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/vv/internal/interceptor"
	"github.com/bluekaki/pkg/vv/internal/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	// ErrNotGrpcContext not a grpc context
	ErrNotGrpcContext = stderr.New("ctx does not contain metadata")

	// ErrNoJournalIDInContext no jouranl_id in ctx
	ErrNoJournalIDInContext = stderr.New("not found jouranl_id in ctx")
)

// JournalID get journal id from context
func JournalID(ctx context.Context) (string, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrNotGrpcContext
	}

	id := meta.Get(interceptor.JournalID)
	if len(id) == 0 {
		return "", ErrNoJournalIDInContext
	}

	return id[0], nil
}

// Error create some error
func Error(c codes.Code, msg string, err errors.Error) error {
	if c == codes.OK && err != nil {
		c = codes.Internal
	}

	s := status.New(c, msg)
	if err != nil {
		s, _ = s.WithDetails(&pb.Stack{Verbose: fmt.Sprintf("%+v", err)})
	}

	return s.Err()
}

func Userinfo(ctx context.Context) interface{} {
	if ctx == nil {
		return nil
	}
	return ctx.Value(interceptor.SessionUserinfo{})
}

func SignatureIdentifier(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	val := ctx.Value(interceptor.SignatureIdentifier{})
	if val == nil {
		return ""
	}

	identifier, ok := val.(string)
	if !ok {
		return ""
	}
	return identifier
}
