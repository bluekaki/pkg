package client

import (
	"strings"
	"time"

	"github.com/bluekaki/pkg/errors"
	"github.com/bluekaki/pkg/vv/internal/configs"
	"github.com/bluekaki/pkg/vv/internal/interceptor"
	"github.com/bluekaki/pkg/vv/proposal"

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

// Option some options for build a conn
type Option func(*option)

type option struct {
	credential      credentials.TransportCredentials
	keepalive       *keepalive.ClientParameters
	resolverBuilder resolver.Builder
	dialTimeout     time.Duration
	signer          proposal.Signer
	projectName     string
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

// WithSigner a handler for do signature
func WithSigner(signer proposal.Signer) Option {
	return func(opt *option) {
		opt.signer = signer
	}
}

// WithProjectName add project name into alert message
func WithProjectName(name string) Option {
	return func(opt *option) {
		opt.projectName = strings.TrimSpace(name)
	}
}

// NewConn create a grpc client conn
func NewConn(endpoint string, logger *zap.Logger, notify proposal.NotifyHandler, options ...Option) (ConnInterface, error) {
	if endpoint = strings.TrimSpace(endpoint); endpoint == "" {
		return nil, errors.New("endpoint required")
	}
	if logger == nil {
		return nil, errors.New("logger required")
	}
	if notify == nil {
		return nil, errors.New("notify required")
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

	dialOptions := []grpc.DialOption{
		grpc.WithResolvers(resolverBuilder),
		grpc.WithTimeout(dialTimeout),
		grpc.WithBlock(),
		grpc.WithMaxMsgSize(configs.MaxMsgSize),
		grpc.WithKeepaliveParams(*kacp),
		grpc.WithUnaryInterceptor(interceptor.UnaryClientInterceptor(logger, notify, opt.signer, opt.projectName)),
		grpc.WithStreamInterceptor(interceptor.StreamClientInterceptor()),
		grpc.WithDefaultServiceConfig(configs.ServiceConfig),
	}

	if opt.credential == nil {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(opt.credential))
	}

	conn, err := grpc.Dial(endpoint, dialOptions...)
	if err != nil {
		return nil, errors.Wrapf(err, "dial %s err", endpoint)
	}

	return &clientConn{conn}, nil
}

// ConnInterface a wrapper for grpc.ClientConnInterface
type ConnInterface interface {
	grpc.ClientConnInterface
	Close() error
	t()
}

type clientConn struct {
	*grpc.ClientConn
}

func (c *clientConn) Close() error {
	return c.ClientConn.Close()
}

func (c *clientConn) t() {}
