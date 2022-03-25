package server

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bluekaki/pkg/vv/internal/configs"
	"github.com/bluekaki/pkg/vv/internal/interceptor"
	"github.com/bluekaki/pkg/vv/proposal"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	defaultEnforcementPolicy = &keepalive.EnforcementPolicy{
		MinTime:             5 * time.Second,
		PermitWithoutStream: true,
	}

	defaultKeepAlive = &keepalive.ServerParameters{
		MaxConnectionIdle: 15 * time.Second,
		Time:              5 * time.Second,
		Timeout:           2 * time.Second,
	}
)

// RegisteAuthorizationValidator userinfo handler for interceptor options.authorization
func RegisteAuthorizationValidator(name string, handler proposal.UserinfoHandler) {
	interceptor.RegisteAuthorizationValidator(name, handler)
}

// RegisteAuthorizationProxyValidator signature handler for interceptor options.authorization_proxy
func RegisteAuthorizationProxyValidator(name string, handler proposal.SignatureHandler) {
	interceptor.RegisteAuthorizationProxyValidator(name, handler)
}

// Option some options for build a server
type Option func(*option)

type option struct {
	credential              credentials.TransportCredentials
	enforcementPolicy       *keepalive.EnforcementPolicy
	keepalive               *keepalive.ServerParameters
	metrics                 func(http.Handler)
	projectName             string
	fds                     []protoreflect.FileDescriptor
	disableMessageValitator bool
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

// WithPrometheus enable prometheus metrics
func WithPrometheus(metrics func(http.Handler)) Option {
	return func(opt *option) {
		opt.metrics = metrics
	}
}

// WithProjectName add project name into alert message
func WithProjectName(name string) Option {
	return func(opt *option) {
		opt.projectName = strings.TrimSpace(name)
	}
}

// WithIgnoreFileDescriptor ignore some method likes grpc/health
func WithIgnoreFileDescriptor(fd protoreflect.FileDescriptor) Option {
	return func(opt *option) {
		if fd != nil {
			opt.fds = append(opt.fds, fd)
		}
	}
}

// WithDisableMessageValitator ignore request message's validator
func WithDisableMessageValitator() Option {
	return func(opt *option) {
		opt.disableMessageValitator = true
	}
}

// RegisterEndpoint the only entrance for register service
type RegisterEndpoint func(server *grpc.Server)

// New create a grpc server
func New(logger *zap.Logger, notify proposal.NotifyHandler, register RegisterEndpoint, options ...Option) GRPCServer {
	if logger == nil {
		panic("logger required")
	}
	if notify == nil {
		panic("notify required")
	}
	if register == nil {
		panic("register required")
	}

	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	enforcementPolicy := defaultEnforcementPolicy
	if opt.enforcementPolicy != nil {
		enforcementPolicy = opt.enforcementPolicy
	}

	keepalive := defaultKeepAlive
	if opt.keepalive != nil {
		keepalive = opt.keepalive
	}

	serverOptions := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(configs.MaxMsgSize),
		grpc.MaxHeaderListSize(configs.MaxMsgSize),
		grpc.KeepaliveEnforcementPolicy(*enforcementPolicy),
		grpc.KeepaliveParams(*keepalive),
		grpc.UnaryInterceptor(interceptor.UnaryServerInterceptor(logger, notify, opt.metrics, opt.projectName, opt.disableMessageValitator)),
		grpc.StreamInterceptor(interceptor.StreamServerInterceptor(logger, notify, opt.metrics, opt.projectName, opt.disableMessageValitator)),
	}

	if opt.credential != nil {
		serverOptions = append(serverOptions, grpc.Creds(opt.credential))
	}

	srv := &grpcServer{
		server: grpc.NewServer(serverOptions...),
	}

	register(srv.server)
	interceptor.ResloveFileDescriptor(interceptor.Server)
	interceptor.IgnoreFileDescriptor(opt.fds)

	return srv
}

// GRPCServer grpc server
type GRPCServer interface {
	Serve(lis net.Listener) error
	GracefulStop()
	t()
}

type grpcServer struct {
	server *grpc.Server
}

func (g *grpcServer) Serve(lis net.Listener) error {
	return g.server.Serve(lis)
}

func (g *grpcServer) GracefulStop() {
	g.server.GracefulStop()
}

func (g *grpcServer) t() {}
