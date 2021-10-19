package server

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/bluekaki/pkg/vv/internal/interceptor"
	"github.com/bluekaki/pkg/vv/pkg/adapter"

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

func RegisteAuthorizationValidator(name string, handler adapter.UserinfoHandler) {
	interceptor.RegisteAuthorizationValidator(name, handler)
}

func RegisteAuthorizationProxyValidator(name string, handler adapter.SignatureHandler) {
	interceptor.RegisteAuthorizationProxyValidator(name, handler)
}

// Option how setup client
type Option func(*option)

type option struct {
	credential        credentials.TransportCredentials
	enforcementPolicy *keepalive.EnforcementPolicy
	keepalive         *keepalive.ServerParameters
	metrics           func(http.Handler)
	projectName       string
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

func WithProjectName(name string) Option {
	return func(opt *option) {
		opt.projectName = strings.TrimSpace(name)
	}
}

type RegisterEndpoint func(server *grpc.Server)

// New create a grpc server
func New(logger *zap.Logger, notify adapter.NotifyHandler, register RegisterEndpoint, options ...Option) GRPCServer {
	if logger == nil {
		panic("logger required")
	}

	opt := new(option)
	for _, f := range options {
		f(opt)
	}

	// if opt.prometheusHandler != nil {
	// 	opt.prometheusHandler(logger)
	// }

	enforcementPolicy := defaultEnforcementPolicy
	if opt.enforcementPolicy != nil {
		enforcementPolicy = opt.enforcementPolicy
	}

	keepalive := defaultKeepAlive
	if opt.keepalive != nil {
		keepalive = opt.keepalive
	}

	serverOptions := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(30 << 20),
		grpc.MaxSendMsgSize(30 << 20),
		grpc.KeepaliveEnforcementPolicy(*enforcementPolicy),
		grpc.KeepaliveParams(*keepalive),
		grpc.UnaryInterceptor(interceptor.UnaryServerInterceptor(logger, notify, opt.metrics, opt.projectName)),
		grpc.StreamInterceptor(interceptor.StreamServerInterceptor()),
	}

	if opt.credential != nil {
		serverOptions = append(serverOptions, grpc.Creds(opt.credential))
	}

	srv := &grpcServer{
		server: grpc.NewServer(serverOptions...),
	}

	register(srv.server)
	interceptor.ResloveFileDescriptor(false)

	return srv
}

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
	g.GracefulStop()
}

func (g *grpcServer) t() {}
