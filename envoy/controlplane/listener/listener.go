package listener

import (
	"time"

	"github.com/bluekaki/pkg/envoy/controlplane/secret"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	connection_limit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/connection_limit/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	defaultMaxConnections = 50000
)

type Option func(*option)

type option struct {
	TLS struct {
		ServerName               string
		Certificate              *tls.TlsCertificate
		RequireClientCertificate *wrapperspb.BoolValue
	}
	MaxConnections uint64
	HTTPManager    *listener.Filter
	TCPManager     *listener.Filter
}

func WithTLS(serverName string, certificate *tls.TlsCertificate, requireClientCertificate bool) Option {
	return func(opt *option) {
		if serverName != "" && certificate != nil {
			opt.TLS.ServerName = serverName
			opt.TLS.Certificate = certificate
			opt.TLS.RequireClientCertificate = wrapperspb.Bool(requireClientCertificate)
		}
	}
}

func WithMaxConnections(maxConnections uint64) Option {
	return func(opt *option) {
		if maxConnections > 0 {
			opt.MaxConnections = maxConnections
		}
	}
}

func WithHTTPManager(manager *listener.Filter) Option {
	return func(opt *option) {
		opt.HTTPManager = manager
	}
}

func WithTCPManager(manager *listener.Filter) Option {
	return func(opt *option) {
		opt.TCPManager = manager
	}
}

func New(name string, port uint32, opts ...Option) *listener.Listener {
	opt := new(option)
	for _, f := range opts {
		f(opt)
	}

	if opt.MaxConnections == 0 {
		opt.MaxConnections = defaultMaxConnections
	}

	return &listener.Listener{
		Name: name,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "::",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
					Ipv4Compat: true,
				},
			},
		},
		StatPrefix: name,
		FilterChains: []*listener.FilterChain{
			{
				FilterChainMatch: func() *listener.FilterChainMatch { // this only for tls
					if opt.TLS.ServerName == "" {
						return nil
					}

					return &listener.FilterChainMatch{
						ServerNames:          []string{opt.TLS.ServerName},
						TransportProtocol:    "tls",
						ApplicationProtocols: nil, // do not set this(some http client donot supports)
					}
				}(),
				Filters: func() (filters []*listener.Filter) {
					filters = append(filters, &listener.Filter{
						Name: "envoy.filters.network.connection_limit",
						ConfigType: &listener.Filter_TypedConfig{
							TypedConfig: func() *anypb.Any {
								config, _ := anypb.New(&connection_limit.ConnectionLimit{
									StatPrefix:     name + "_connection_limit",
									MaxConnections: &wrappers.UInt64Value{Value: opt.MaxConnections},
									Delay:          durationpb.New(time.Second),
								})

								return config
							}(),
						},
					})

					// the HCM and tcp_proxy filters should not be used together
					switch {
					case opt.HTTPManager != nil:
						filters = append(filters, opt.HTTPManager)

					case opt.TCPManager != nil:
						filters = append(filters, opt.TCPManager)
					}

					return
				}(),
				TransportSocket: func() *core.TransportSocket {
					if opt.TLS.ServerName == "" {
						return nil
					}

					return &core.TransportSocket{
						Name: wellknown.TransportSocketTLS,
						ConfigType: &core.TransportSocket_TypedConfig{
							TypedConfig: func() *anypb.Any {
								config, _ := anypb.New(&tls.DownstreamTlsContext{
									CommonTlsContext: &tls.CommonTlsContext{
										TlsParams:       secret.NewTlsParameters(),
										TlsCertificates: []*tls.TlsCertificate{opt.TLS.Certificate},
									},
									RequireClientCertificate: opt.TLS.RequireClientCertificate,
								})

								return config
							}(),
						},
					}
				}(),
				TransportSocketConnectTimeout: func() *durationpb.Duration {
					// this must use with tls, if not will has higher priority than idle timeout.
					if opt.TLS.ServerName == "" {
						return nil
					}

					return durationpb.New(time.Second * 2)
				}(),
			},
		},
		ListenerFilters: func() (filters []*listener.ListenerFilter) { // used by filterchains
			if opt.TLS.ServerName != "" {
				filters = append(filters, &listener.ListenerFilter{Name: wellknown.TLSInspector})
			}
			if opt.HTTPManager != nil {
				filters = append(filters, &listener.ListenerFilter{Name: wellknown.HTTPInspector})
			}

			return
		}(),
		ListenerFiltersTimeout: durationpb.New(time.Second * 2),
	}
}
