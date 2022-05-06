package http

import (
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	cors "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/cors/v3"
	local_ratelimit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	http_wasm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/wasm/v3"
	http_connection_manager "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	wasm "github.com/envoyproxy/go-control-plane/envoy/extensions/wasm/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Option func(*option)

type option struct {
	WASMFilter *http_connection_manager.HttpFilter
	ServerName string
	Via        string
}

func WithWASMFilter(url, digest string) Option {
	return func(opt *option) {
		opt.WASMFilter = &http_connection_manager.HttpFilter{
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

func WithServerName(name string) Option {
	return func(opt *option) {
		opt.ServerName = name
	}
}

func WithVia(via string) Option {
	return func(opt *option) {
		opt.Via = via
	}
}

func New(name string, opts ...Option) *listener.Filter {
	opt := new(option)
	for _, f := range opts {
		f(opt)
	}

	return &listener.Filter{
		Name: wellknown.HTTPConnectionManager,
		ConfigType: &listener.Filter_TypedConfig{
			TypedConfig: func() *anypb.Any {
				config, _ := anypb.New(&http_connection_manager.HttpConnectionManager{
					CodecType:  http_connection_manager.HttpConnectionManager_AUTO,
					StatPrefix: name,
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

						if opt.WASMFilter != nil {
							filters = append(filters, opt.WASMFilter)
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
					ServerName:                   opt.ServerName,
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
	}
}
