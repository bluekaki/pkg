package client

import (
	"time"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/vv/internal/configs"
	"github.com/bluekaki/pkg/vv/internal/interceptor"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/dns"
)

var (
	defaultKeepAlive = &keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             2 * time.Second,
		PermitWithoutStream: true,
	}

	defaultResolverBuilder = dns.NewBuilder()

	defaultDialTimeout = time.Second * 2
)

// Sign signs the message
type Sign = interceptor.Sign

// Option how setup client
type Option func(*option)

type option struct {
	credential      credentials.TransportCredentials
	keepalive       *keepalive.ClientParameters
	resolverBuilder resolver.Builder
	dialTimeout     time.Duration
	sign            Sign
	notifyHandler   interceptor.NotifyHandler
}

// WithCredential setup credential for tls
func WithCredential(credential credentials.TransportCredentials) Option {
	return func(opt *option) {
		opt.credential = credential
	}
}

// WithKeepAlive setup keepalive parameters
func WithKeepAlive(keepalive *keepalive.ClientParameters) Option {
	return func(opt *option) {
		opt.keepalive = keepalive
	}
}

// WithResolverBuilder setup resolver builder
func WithResolverBuilder(builder resolver.Builder) Option {
	return func(opt *option) {
		opt.resolverBuilder = builder
	}
}

// WithDialTimeout setup the dial timeout
func WithDialTimeout(timeout time.Duration) Option {
	return func(opt *option) {
		opt.dialTimeout = timeout
	}
}

// WithSign setup the signature handler
func WithSign(sign Sign) Option {
	return func(opt *option) {
		opt.sign = sign
	}
}

// WithNotifyHandler notify when got panic
func WithNotifyHandler(handler interceptor.NotifyHandler) Option {
	return func(opt *option) {
		opt.notifyHandler = handler
	}
}

// New create a grpc client conn
func New(logger *zap.Logger, endpoint string, options ...Option) (*grpc.ClientConn, error) {
	if logger == nil {
		panic("logger required")
	}

	if endpoint == "" {
		return nil, errors.New("endpoint required")
	}

	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	kacp := defaultKeepAlive
	if opt.keepalive != nil {
		kacp = opt.keepalive
	}

	resolverBuilder := defaultResolverBuilder
	if opt.resolverBuilder != nil {
		resolverBuilder = opt.resolverBuilder
	}

	dialTimeout := defaultDialTimeout
	if opt.dialTimeout > 0 {
		dialTimeout = opt.dialTimeout
	}

	clientInterceptor := interceptor.NewClientInterceptor(opt.sign, logger, opt.notifyHandler)

	dialOptions := []grpc.DialOption{
		grpc.WithResolvers(resolverBuilder),
		grpc.WithTimeout(dialTimeout),
		grpc.WithBlock(),
		grpc.WithMaxMsgSize(20 << 20),
		grpc.WithKeepaliveParams(*kacp),
		grpc.WithUnaryInterceptor(clientInterceptor.UnaryInterceptor),
		grpc.WithStreamInterceptor(clientInterceptor.StreamInterceptor),
		grpc.WithDefaultServiceConfig(configs.ServiceConfig),
	}

	if opt.credential == nil {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(opt.credential))
	}

	conn, err := grpc.Dial(endpoint, dialOptions...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return conn, nil
}
