package gateway

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bluekaki/pkg/id"
	"github.com/bluekaki/pkg/vv/internal/configs"
	"github.com/bluekaki/pkg/vv/internal/interceptor"
	"github.com/bluekaki/pkg/vv/internal/pkg/marshaler"
	"github.com/bluekaki/pkg/vv/internal/pkg/multipart"
	"github.com/bluekaki/pkg/vv/proposal"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver/dns"
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

// RegisteWhitelistingValidator whiteling handler for interceptor options.whitelisting
func RegisteWhitelistingValidator(name string, handler proposal.WhitelistingHandler) {
	interceptor.RegisteWhitelistingValidator(name, handler)
}

// Option some options for build a gateway
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

// WithProjectName add project name into alert message
func WithProjectName(name string) Option {
	return func(opt *option) {
		opt.projectName = strings.TrimSpace(name)
	}
}

// RegisterEndpoint the only entrance for register backend endpoints
type RegisterEndpoint func(mux *runtime.ServeMux, opts []grpc.DialOption) error

// NewCorsHandler create a cors http handler
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

	jsonPbMarshaler := marshaler.NewJSONPbMarshaler()
	fromDataMarshaler := marshaler.NewFromDataMarshaler()

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(runtime.DefaultHeaderMatcher),
		runtime.WithOutgoingHeaderMatcher(runtime.DefaultHeaderMatcher),
		runtime.WithMetadata(annotator(logger)),
		runtime.WithErrorHandler(runtime.DefaultHTTPErrorHandler),
		runtime.WithStreamErrorHandler(runtime.DefaultStreamErrorHandler),
		runtime.WithRoutingErrorHandler(runtime.DefaultRoutingErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaler.NewWildcardMarshaler()),
		runtime.WithMarshalerOption("application/x-www-form-urlencoded", jsonPbMarshaler),
		runtime.WithMarshalerOption("application/json", jsonPbMarshaler),
		runtime.WithMarshalerOption("multipart/form-data", fromDataMarshaler),
	)

	dialOptions := []grpc.DialOption{
		grpc.WithResolvers(dns.NewBuilder()),
		grpc.WithTimeout(dialTimeout),
		grpc.WithBlock(),
		grpc.WithMaxMsgSize(configs.MaxMsgSize),
		grpc.WithMaxHeaderListSize(configs.MaxMsgSize),
		grpc.WithKeepaliveParams(*kacp),
		grpc.WithUnaryInterceptor(interceptor.UnaryGatewayInterceptor(logger, notify, opt.metrics, opt.projectName)),
		grpc.WithStreamInterceptor(interceptor.StreamGatewayInterceptor(logger, notify, opt.metrics, opt.projectName)),
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
	interceptor.ResloveFileDescriptor(interceptor.Gateway)

	return cors.AllowAll().Handler(mux)
}

func annotator(logger *zap.Logger) func(ctx context.Context, req *http.Request) metadata.MD {

	return func(ctx context.Context, req *http.Request) metadata.MD {
		journalID := req.Header.Get(interceptor.JournalID)
		if journalID == "" {
			journalID = id.JournalID()
		}

		body, octet, err := multipart.Parse(req)
		if err != nil {
			logger.Error(fmt.Sprintf("parse multipart err [Journal-Id: %s]", journalID), zap.Error(err))
		}

		req.Body = io.NopCloser(bytes.NewBuffer(body)) // re-construct req body

		return metadata.Pairs(
			interceptor.JournalID, journalID,
			interceptor.Authorization, req.Header.Get(interceptor.Authorization),
			interceptor.AuthorizationProxy, req.Header.Get(interceptor.AuthorizationProxy),
			interceptor.Date, req.Header.Get(interceptor.Date),
			interceptor.Method, req.Method,
			interceptor.URI, req.RequestURI,
			interceptor.Body, func() string {
				if octet {
					return base64.StdEncoding.EncodeToString(body)
				}
				return string(body)
			}(),
			interceptor.XForwardedFor, req.Header.Get(interceptor.XForwardedFor),
			interceptor.XForwardedHost, req.Header.Get(interceptor.XForwardedHost),
			interceptor.OctetStream, func() string {
				if octet {
					return "base64"
				}
				return ""
			}(),
		)
	}
}
