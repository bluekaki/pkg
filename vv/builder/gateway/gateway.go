package gateway

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/bluekaki/pkg/id"
	"github.com/bluekaki/pkg/vv/internal/configs"
	"github.com/bluekaki/pkg/vv/internal/interceptor"
	"github.com/bluekaki/pkg/vv/proposal"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver/dns"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	defaultKeepAlive = &keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             2 * time.Second,
		PermitWithoutStream: true,
	}

	defaultDialTimeout = time.Second * 2
)

func init() {
	runtime.DefaultContextTimeout = time.Second * 30
}

func RegisteWhitelistingValidator(name string, handler proposal.WhitelistingHandler) {
	interceptor.RegisteWhitelistingValidator(name, handler)
}

// Option how setup client
type Option func(*option)

type option struct {
	credential  credentials.TransportCredentials
	keepalive   *keepalive.ClientParameters
	dialTimeout time.Duration
	metrics     func(http.Handler)
	projectName string
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

// WithDialTimeout setup the dial timeout
func WithDialTimeout(timeout time.Duration) Option {
	return func(opt *option) {
		opt.dialTimeout = timeout
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

type RegisterEndpoint func(mux *runtime.ServeMux, opts []grpc.DialOption) error

func NewCorsHandler(logger *zap.Logger, notify proposal.NotifyHandler, register RegisterEndpoint, options ...Option) http.Handler {
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

	kacp := defaultKeepAlive
	if opt.keepalive != nil {
		kacp = opt.keepalive
	}

	dialTimeout := defaultDialTimeout
	if opt.dialTimeout > 0 {
		dialTimeout = opt.dialTimeout
	}

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(runtime.DefaultHeaderMatcher),
		runtime.WithOutgoingHeaderMatcher(runtime.DefaultHeaderMatcher),
		runtime.WithMetadata(annotator),
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler), // TODO convert http code
		runtime.WithStreamErrorHandler(runtime.DefaultStreamErrorHandler),
		runtime.WithRoutingErrorHandler(runtime.DefaultRoutingErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
	)

	dialOptions := []grpc.DialOption{
		grpc.WithResolvers(dns.NewBuilder()),
		grpc.WithTimeout(dialTimeout),
		grpc.WithBlock(),
		grpc.WithMaxMsgSize(configs.MaxMsgSize),
		grpc.WithKeepaliveParams(*kacp),
		grpc.WithUnaryInterceptor(interceptor.UnaryGatewayInterceptor(logger, notify, opt.metrics, opt.projectName)),
		grpc.WithDefaultServiceConfig(configs.ServiceConfig),
	}

	if opt.credential == nil {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(opt.credential))
	}

	if err := register(mux, dialOptions); err != nil {
		panic(err)
	}
	interceptor.ResloveFileDescriptor(true)

	return cors.AllowAll().Handler(mux)
}

func annotator(ctx context.Context, req *http.Request) metadata.MD {
	body, _ := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // re-construct req body

	journalID := req.Header.Get(interceptor.JournalID)
	if journalID == "" {
		journalID = id.JournalID()
	}

	return metadata.Pairs(
		interceptor.JournalID, journalID,
		interceptor.Authorization, req.Header.Get(interceptor.Authorization),
		interceptor.AuthorizationProxy, req.Header.Get(interceptor.AuthorizationProxy),
		interceptor.Date, req.Header.Get(interceptor.Date),
		interceptor.Method, req.Method,
		interceptor.URI, req.RequestURI,
		interceptor.Body, string(body),
		interceptor.XForwardedFor, req.Header.Get(interceptor.XForwardedFor),
		interceptor.XForwardedHost, req.Header.Get(interceptor.XForwardedHost),
	)
}
