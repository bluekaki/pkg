package gateway

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"github.com/byepichi/pkg/minami58"
	"github.com/byepichi/pkg/vv/internal/configs"
	"github.com/byepichi/pkg/vv/internal/interceptor"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver/dns"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

var (
	defaultKeepAlive = &keepalive.ClientParameters{
		Time:                10 * time.Second,
		Timeout:             time.Second,
		PermitWithoutStream: true,
	}

	defaultDialTimeout = time.Second * 2
)

func init() {
	runtime.DefaultContextTimeout = time.Second * 10
}

// Option how setup client
type Option func(*option)

type option struct {
	credential     credentials.TransportCredentials
	keepalive      *keepalive.ClientParameters
	dialTimeout    time.Duration
	marshalJournal bool
	notifyHandler  func(desc, err, stack, journalID string)
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

// WithMarshalJournal marshal journal to json string
func WithMarshalJournal() Option {
	return func(opt *option) {
		opt.marshalJournal = true
	}
}

// WithNotifyHandler notify when got panic
func WithNotifyHandler(handler func(desc, err, stack, journalID string)) Option {
	return func(opt *option) {
		opt.notifyHandler = handler
	}
}

// New create grpc-gateway server mux, and grpc dial options.
func New(logger *zap.Logger, options ...Option) (*runtime.ServeMux, []grpc.DialOption) {
	if logger == nil {
		panic("logger required")
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

	// copy from runtime.DefaultHTTPErrorHandler
	errHandler := func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
		// return Internal when Marshal failed
		const fallback = `{"code": 13, "message": "failed to marshal error message"}`

		s := status.Convert(err)
		pb := s.Proto()

		w.Header().Del("Trailer")
		w.Header().Del("Transfer-Encoding")

		contentType := marshaler.ContentType(pb)
		w.Header().Set("Content-Type", contentType)

		buf, merr := marshaler.Marshal(pb)
		if merr != nil {
			grpclog.Infof("Failed to marshal error message %q: %v", s, merr)
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := io.WriteString(w, fallback); err != nil {
				grpclog.Infof("Failed to write response: %v", err)
			}
			return
		}

		md, ok := runtime.ServerMetadataFromContext(ctx)
		if !ok {
			grpclog.Infof("Failed to extract ServerMetadata from context")
		}

		for k, vs := range md.HeaderMD {
			if h, ok := runtime.DefaultHeaderMatcher(k); ok {
				for _, v := range vs {
					w.Header().Add(h, v)
				}
			}
		}

		// RFC 7230 https://tools.ietf.org/html/rfc7230#section-4.1.2
		// Unless the request includes a TE header field indicating "trailers"
		// is acceptable, as described in Section 4.3, a server SHOULD NOT
		// generate trailer fields that it believes are necessary for the user
		// agent to receive.
		var wantsTrailers bool

		if te := r.Header.Get("TE"); strings.Contains(strings.ToLower(te), "trailers") {
			wantsTrailers = true
			for k := range md.TrailerMD {
				tKey := textproto.CanonicalMIMEHeaderKey(fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k))
				w.Header().Add("Trailer", tKey)
			}
			w.Header().Set("Transfer-Encoding", "chunked")
		}

		st := runtime.HTTPStatusFromCode(s.Code())
		if code := strconv.Itoa(int(s.Code())); len(code) == 6 {
			hst, _ := strconv.Atoi(code[:3])
			if http.StatusText(hst) != "" {
				st = hst
			}
		}

		w.WriteHeader(st)
		if _, err := w.Write(buf); err != nil {
			grpclog.Infof("Failed to write response: %v", err)
		}

		if wantsTrailers {
			for k, vs := range md.TrailerMD {
				tKey := fmt.Sprintf("%s%s", runtime.MetadataTrailerPrefix, k)
				for _, v := range vs {
					w.Header().Add(tKey, v)
				}
			}
		}
	}

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(runtime.DefaultHeaderMatcher),
		runtime.WithOutgoingHeaderMatcher(runtime.DefaultHeaderMatcher),
		runtime.WithMetadata(annotator),
		runtime.WithErrorHandler(errHandler),
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

	gatewayInterceptor := interceptor.NewGatewayInterceptor(logger, opt.marshalJournal, opt.notifyHandler)

	dialOptions := []grpc.DialOption{
		grpc.WithResolvers(dns.NewBuilder()),
		grpc.WithTimeout(dialTimeout),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(*kacp),
		grpc.WithUnaryInterceptor(gatewayInterceptor.UnaryInterceptor),
		grpc.WithDefaultServiceConfig(configs.ServiceConfig),
	}

	if opt.credential == nil {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	} else {
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(opt.credential))
	}

	return mux, dialOptions
}

func annotator(ctx context.Context, req *http.Request) metadata.MD {
	body, _ := ioutil.ReadAll(req.Body)
	req.Body = ioutil.NopCloser(bytes.NewBuffer(body)) // re-construct req body

	journalID := req.Header.Get(interceptor.JournalID)
	if journalID == "" {
		nonce := make([]byte, 16)
		io.ReadFull(rand.Reader, nonce)
		journalID = string(minami58.Encode(nonce))
	}

	return metadata.Pairs(
		interceptor.JournalID, journalID,
		interceptor.Authorization, req.Header.Get(interceptor.Authorization),
		interceptor.ProxyAuthorization, req.Header.Get(interceptor.ProxyAuthorization),
		interceptor.MixAuthorization, req.Header.Get(interceptor.MixAuthorization),
		interceptor.Date, req.Header.Get(interceptor.Date),
		interceptor.Method, req.Method,
		interceptor.URI, req.RequestURI,
		interceptor.Body, string(body), // TODO unsafe
		interceptor.XForwardedFor, req.Header.Get(interceptor.XForwardedFor),
		interceptor.XForwardedHost, req.Header.Get(interceptor.XForwardedHost),
	)
}
