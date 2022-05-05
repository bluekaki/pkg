package listener

import (
	"time"

	"github.com/bluekaki/pkg/envoy/controlplane/secret"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	cors "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cors/v3"
	local_ratelimit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	http_wasm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/wasm/v3"
	connection_limit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/connection_limit/v3"
	http_connection_manager "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	wasm "github.com/envoyproxy/go-control-plane/envoy/extensions/wasm/v3"
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
	WASM           struct {
		HTTP *http_connection_manager.HttpFilter
	}
	Via string
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

func WithHttpWasmFilter(url, digest string) Option {
	return func(opt *option) {
		opt.WASM.HTTP = &http_connection_manager.HttpFilter{
			Name: "envoy.filters.http.wasm",
			ConfigType: &http_connection_manager.HttpFilter_TypedConfig{
				TypedConfig: func() *anypb.Any {
					config, _ := anypb.New(&http_wasm.Wasm{
						Config: &wasm.PluginConfig{
							Name:   "plugin_http_filter",
							RootId: "wasm_http_filter",
							Vm: &wasm.PluginConfig_VmConfig{
								VmConfig: &wasm.VmConfig{
									VmId:    "vm_http_filter",
									Runtime: "envoy.wasm.runtime.v8",
									Code: &core.AsyncDataSource{
										Specifier: &core.AsyncDataSource_Remote{
											Remote: &core.RemoteDataSource{
												HttpUri: &core.HttpUri{
													Uri: url,
													HttpUpstreamType: &core.HttpUri_Cluster{
														Cluster: "wasm_cluster",
													},
													Timeout: durationpb.New(time.Second * 20),
												},
												Sha256: digest,
												RetryPolicy: &core.RetryPolicy{
													RetryBackOff: &core.BackoffStrategy{
														BaseInterval: durationpb.New(time.Second),
													},
													NumRetries: wrapperspb.UInt32(10),
												},
											},
										},
									},
								},
							},
						},
					})

					return config
				}(),
			},
		}
	}
}

func WithVia(via string) Option {
	return func(opt *option) {
		opt.Via = via
	}
}

func NewHTTP_GRPC(name string, port uint32, opts ...Option) *listener.Listener {
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

					filters = append(filters, &listener.Filter{
						Name: wellknown.HTTPConnectionManager,
						ConfigType: &listener.Filter_TypedConfig{
							TypedConfig: func() *anypb.Any {
								config, _ := anypb.New(&http_connection_manager.HttpConnectionManager{
									CodecType:  http_connection_manager.HttpConnectionManager_AUTO,
									StatPrefix: name + "_http(2)",
									RouteSpecifier: &http_connection_manager.HttpConnectionManager_Rds{
										Rds: &http_connection_manager.Rds{
											ConfigSource: &core.ConfigSource{
												ConfigSourceSpecifier: &core.ConfigSource_ApiConfigSource{
													ApiConfigSource: &core.ApiConfigSource{
														ApiType:             core.ApiConfigSource_GRPC,
														TransportApiVersion: core.ApiVersion_V3,
														GrpcServices: []*core.GrpcService{{
															TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
																EnvoyGrpc: &core.GrpcService_EnvoyGrpc{
																	ClusterName: "ads_cluster",
																},
															},
															Timeout: durationpb.New(time.Second * 3),
														}},
													},
												},
												InitialFetchTimeout: durationpb.New(time.Second * 5),
												ResourceApiVersion:  core.ApiVersion_V3,
											},
											RouteConfigName: "http_grpc_rds_config",
										},
									},
									HttpFilters: func() (filters []*http_connection_manager.HttpFilter) {
										filters = append(filters, &http_connection_manager.HttpFilter{
											// this is necessary for router's local_ratelimit
											Name: "envoy.filters.http.local_ratelimit",
											ConfigType: &http_connection_manager.HttpFilter_TypedConfig{
												TypedConfig: func() *anypb.Any {
													config, _ := anypb.New(&local_ratelimit.LocalRateLimit{
														StatPrefix: name + "_local_ratelimit",
													})

													return config
												}(),
											},
										})

										filters = append(filters, &http_connection_manager.HttpFilter{
											Name: "envoy.filters.http.cors",
											ConfigType: &http_connection_manager.HttpFilter_TypedConfig{
												TypedConfig: func() *anypb.Any {
													config, _ := anypb.New(&cors.Cors{})

													return config
												}(),
											},
										})

										if opt.WASM.HTTP != nil {
											filters = append(filters, opt.WASM.HTTP)
										}

										filters = append(filters, &http_connection_manager.HttpFilter{
											Name: wellknown.Router,
										})

										return
									}(),
									Tracing: nil, // TODO
									HttpProtocolOptions: &core.Http1ProtocolOptions{
										EnableTrailers: true,
									},
									Http2ProtocolOptions: &core.Http2ProtocolOptions{
										ConnectionKeepalive: &core.KeepaliveSettings{
											Interval:               durationpb.New(time.Second * 3),
											Timeout:                durationpb.New(time.Second * 2),
											ConnectionIdleInterval: durationpb.New(time.Second * 2),
										},
									},
									StreamIdleTimeout:            nil, // overridable by the route-level idle_timeout
									RequestTimeout:               durationpb.New(time.Second * 10),
									RequestHeadersTimeout:        durationpb.New(time.Second * 2),
									DrainTimeout:                 durationpb.New(time.Second * 10),
									DelayedCloseTimeout:          durationpb.New(time.Second),
									AccessLog:                    nil, // recored in wasm
									UseRemoteAddress:             wrapperspb.Bool(true),
									Via:                          opt.Via,
									GenerateRequestId:            wrapperspb.Bool(false), // gen by wasm
									PreserveExternalRequestId:    false,                  // gen by wasm
									AlwaysSetRequestIdInResponse: false,                  // gen by wasm
									NormalizePath:                wrapperspb.Bool(true),  // return 400 if illegal
									StripMatchingHostPort:        true,
									StripTrailingHostDot:         true,
								})

								return config
							}(),
						},
					})

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
		ListenerFilters: []*listener.ListenerFilter{
			{Name: wellknown.TLSInspector},
			{Name: wellknown.HTTPInspector},
		},
		ListenerFiltersTimeout: durationpb.New(time.Second * 2),
	}
}
