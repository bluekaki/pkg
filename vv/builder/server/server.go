package server

import (
	"time"

	"github.com/bluekaki/pkg/vv/internal/interceptor"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	_ "google.golang.org/grpc/encoding/gzip" // enable gzip
	"google.golang.org/grpc/keepalive"
)

var (
	defaultEnforcementPolicy = &keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}

	defaultKeepAlive = &keepalive.ServerParameters{
		MaxConnectionIdle:     15 * time.Second,
		MaxConnectionAge:      30 * time.Second,
		MaxConnectionAgeGrace: 5 * time.Second,
		Time:                  5 * time.Second,
		Timeout:               2 * time.Second,
	}
)

// Option how setup client
type Option func(*option)

type option struct {
	credential        credentials.TransportCredentials
	enforcementPolicy *keepalive.EnforcementPolicy
	keepalive         *keepalive.ServerParameters
	prometheusHandler func(*zap.Logger)
	notifyHandler     interceptor.NotifyHandler
}

// WithCredential setup credential for tls
func WithCredential(credential credentials.TransportCredentials) Option {
	return func(opt *option) {
		opt.credential = credential
	}
}

// WithEnforcementPolicy setup enforcement policy
func WithEnforcementPolicy(enforcementPolicy *keepalive.EnforcementPolicy) Option {
	return func(opt *option) {
		opt.enforcementPolicy = enforcementPolicy
	}
}

// WithKeepAlive setup keepalive parameters
func WithKeepAlive(keepalive *keepalive.ServerParameters) Option {
	return func(opt *option) {
		opt.keepalive = keepalive
	}
}

// WithNotifyHandler notify when got panic
func WithNotifyHandler(handler interceptor.NotifyHandler) Option {
	return func(opt *option) {
		opt.notifyHandler = handler
	}
}

// New create a grpc server
func New(logger *zap.Logger, options ...Option) *grpc.Server {
	if logger == nil {
		panic("logger required")
	}

	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	if opt.prometheusHandler != nil {
		opt.prometheusHandler(logger)
	}

	enforcementPolicy := defaultEnforcementPolicy
	if opt.enforcementPolicy != nil {
		enforcementPolicy = opt.enforcementPolicy
	}

	keepalive := defaultKeepAlive
	if opt.keepalive != nil {
		keepalive = opt.keepalive
	}

	serverInterceptor := interceptor.NewServerInterceptor(logger, opt.prometheusHandler != nil, opt.notifyHandler)

	serverOptions := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(20 << 20),
		grpc.MaxSendMsgSize(20 << 20),
		grpc.KeepaliveEnforcementPolicy(*enforcementPolicy),
		grpc.KeepaliveParams(*keepalive),
		grpc.UnaryInterceptor(serverInterceptor.UnaryInterceptor),
		grpc.StreamInterceptor(serverInterceptor.StreamInterceptor),
	}

	if opt.credential != nil {
		serverOptions = append(serverOptions, grpc.Creds(opt.credential))
	}

	return grpc.NewServer(serverOptions...)
}
