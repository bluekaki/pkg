package cluster

import (
	"time"

	"github.com/bluekaki/pkg/envoy/controlplane/secret"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Target struct {
	Host   string
	Port   uint32
	TLS    bool
	Weight uint32
}
type Option func(*option)

type option struct {
	HealthChecker struct {
		HttpHealthCheck *core.HealthCheck_HttpHealthCheck_
		GrpcHealthCheck *core.HealthCheck_GrpcHealthCheck_
		TcpHealthCheck  *core.HealthCheck_TcpHealthCheck_
	}
	UpstreamProtocolOptions *http.HttpProtocolOptions_ExplicitHttpConfig_
}

func WithHTTPHealthCheck(path string, headers map[string]string, expectedServiceName string) Option {
	return func(opt *option) {
		opt.HealthChecker.HttpHealthCheck = &core.HealthCheck_HttpHealthCheck_{
			HttpHealthCheck: &core.HealthCheck_HttpHealthCheck{
				Path: path,
				RequestHeadersToAdd: func() (_headers []*core.HeaderValueOption) {
					for k, v := range headers {
						_headers = append(_headers, &core.HeaderValueOption{
							Header: &core.HeaderValue{Key: k, Value: v},
							Append: &wrappers.BoolValue{Value: false},
						})
					}

					return
				}(),
				CodecClientType: _type.CodecClientType_HTTP1,
				ServiceNameMatcher: func() *matcher.StringMatcher {
					if expectedServiceName == "" {
						return nil
					}

					return &matcher.StringMatcher{
						MatchPattern: &matcher.StringMatcher_Exact{
							Exact: expectedServiceName, // resp.header of x-envoy-upstream-healthchecked-cluster
						},
					}
				}(),
			},
		}
	}
}

func WithGRPCHealthCheck(serviceName string) Option {
	return func(opt *option) {
		opt.HealthChecker.GrpcHealthCheck = &core.HealthCheck_GrpcHealthCheck_{
			GrpcHealthCheck: &core.HealthCheck_GrpcHealthCheck{
				ServiceName: serviceName,
			},
		}
	}
}

func WithTCPHealthCheck() Option {
	return func(opt *option) {
		opt.HealthChecker.TcpHealthCheck = &core.HealthCheck_TcpHealthCheck_{
			TcpHealthCheck: &core.HealthCheck_TcpHealthCheck{},
		}
	}
}

func WithHTTP1() Option {
	return func(opt *option) {
		opt.UpstreamProtocolOptions = &http.HttpProtocolOptions_ExplicitHttpConfig_{
			ExplicitHttpConfig: &http.HttpProtocolOptions_ExplicitHttpConfig{
				ProtocolConfig: &http.HttpProtocolOptions_ExplicitHttpConfig_HttpProtocolOptions{
					HttpProtocolOptions: &core.Http1ProtocolOptions{
						EnableTrailers: true,
					},
				},
			},
		}
	}
}

func WithHTTP2() Option {
	return func(opt *option) {
		opt.UpstreamProtocolOptions = &http.HttpProtocolOptions_ExplicitHttpConfig_{
			ExplicitHttpConfig: &http.HttpProtocolOptions_ExplicitHttpConfig{
				ProtocolConfig: &http.HttpProtocolOptions_ExplicitHttpConfig_Http2ProtocolOptions{
					Http2ProtocolOptions: &core.Http2ProtocolOptions{
						ConnectionKeepalive: &core.KeepaliveSettings{
							Interval:               durationpb.New(time.Second * 3),
							Timeout:                durationpb.New(time.Second * 2),
							ConnectionIdleInterval: durationpb.New(time.Second * 2),
						},
					},
				},
			},
		}
	}
}

func New(name string, targets []*Target, opts ...Option) *cluster.Cluster {
	opt := new(option)
	for _, f := range opts {
		f(opt)
	}

	if opt.HealthChecker.GrpcHealthCheck == nil &&
		opt.HealthChecker.HttpHealthCheck == nil &&
		opt.HealthChecker.TcpHealthCheck == nil {
		panic("health checker required")
	}

	if opt.UpstreamProtocolOptions == nil {
		WithHTTP1()(opt)
	}

	return &cluster.Cluster{
		TransportSocketMatches: func() (transports []*cluster.Cluster_TransportSocketMatch) {
			for _, target := range targets {
				if !target.TLS {
					continue
				}

				match, err := structpb.NewStruct(map[string]interface{}{
					target.Host: true,
				})
				if err != nil {
					panic(err)
				}

				transports = append(transports, &cluster.Cluster_TransportSocketMatch{
					Name:  target.Host + "_transportsocket",
					Match: match,
					TransportSocket: &core.TransportSocket{
						Name: wellknown.TransportSocketTLS,
						ConfigType: &core.TransportSocket_TypedConfig{
							TypedConfig: func() *anypb.Any {
								config, _ := anypb.New(&tls.UpstreamTlsContext{
									CommonTlsContext: &tls.CommonTlsContext{
										TlsParams: secret.NewTlsParameters(),
									},
									Sni: target.Host,
								})

								return config
							}(),
						},
					},
				})
			}

			return
		}(),
		Name:                 name,
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STRICT_DNS},
		ConnectTimeout:       durationpb.New(time.Second * 2),
		LbPolicy:             cluster.Cluster_RING_HASH,
		LoadAssignment: &endpoint.ClusterLoadAssignment{
			ClusterName: name,
			Endpoints: []*endpoint.LocalityLbEndpoints{{
				LbEndpoints: func() (endpoints []*endpoint.LbEndpoint) {
					for _, target := range targets {
						if target.Weight == 0 {
							target.Weight = 1
						}

						match, err := structpb.NewStruct(map[string]interface{}{
							target.Host: true,
						})
						if err != nil {
							panic(err)
						}

						endpoints = append(endpoints, &endpoint.LbEndpoint{
							HostIdentifier: &endpoint.LbEndpoint_Endpoint{
								Endpoint: &endpoint.Endpoint{
									Address: &core.Address{
										Address: &core.Address_SocketAddress{
											SocketAddress: &core.SocketAddress{
												Protocol: core.SocketAddress_TCP,
												Address:  target.Host,
												PortSpecifier: &core.SocketAddress_PortValue{
													PortValue: target.Port,
												},
											},
										},
									},
									HealthCheckConfig: &endpoint.Endpoint_HealthCheckConfig{
										Hostname: target.Host,
									},
									Hostname: target.Host,
								},
							},
							Metadata: &core.Metadata{
								FilterMetadata: map[string]*_struct.Struct{
									"envoy.transport_socket_match": match,
								},
							},
							LoadBalancingWeight: wrapperspb.UInt32(target.Weight),
						})
					}

					return
				}(),
			}},
		},
		HealthChecks: func() []*core.HealthCheck { // 503 Service Unavailable
			check := &core.HealthCheck{
				Timeout:                  durationpb.New(time.Millisecond * 1500),
				Interval:                 durationpb.New(time.Second * 2),
				InitialJitter:            durationpb.New(time.Second * 2),
				UnhealthyThreshold:       &wrappers.UInt32Value{Value: 2},
				HealthyThreshold:         &wrappers.UInt32Value{Value: 2},
				HealthChecker:            nil,
				NoTrafficInterval:        durationpb.New(time.Second * 2),
				NoTrafficHealthyInterval: durationpb.New(time.Second * 2),
				// EventLogPath:                 os.Stderr.Name(),
				// AlwaysLogHealthCheckFailures: true,
			}

			if opt.HealthChecker.GrpcHealthCheck != nil {
				check.HealthChecker = opt.HealthChecker.GrpcHealthCheck
			} else if opt.HealthChecker.HttpHealthCheck != nil {
				check.HealthChecker = opt.HealthChecker.HttpHealthCheck
			} else if opt.HealthChecker.TcpHealthCheck != nil {
				check.HealthChecker = opt.HealthChecker.TcpHealthCheck
			}

			return []*core.HealthCheck{check}
		}(),
		CircuitBreakers: &cluster.CircuitBreakers{ // 503 Service Unavailable
			Thresholds: []*cluster.CircuitBreakers_Thresholds{{
				MaxConnections:     &wrappers.UInt32Value{Value: 1024},
				MaxPendingRequests: &wrappers.UInt32Value{Value: 1024},
				MaxRequests:        &wrappers.UInt32Value{Value: 1024},
				MaxRetries:         &wrappers.UInt32Value{Value: 3},
			}},
		},
		TypedExtensionProtocolOptions: map[string]*anypb.Any{
			"envoy.extensions.upstreams.http.v3.HttpProtocolOptions": func() *anypb.Any {
				config, _ := anypb.New(&http.HttpProtocolOptions{
					UpstreamProtocolOptions: opt.UpstreamProtocolOptions,
				})

				return config
			}(),
		},
		LbConfig: &cluster.Cluster_RingHashLbConfig_{
			RingHashLbConfig: &cluster.Cluster_RingHashLbConfig{
				HashFunction: cluster.Cluster_RingHashLbConfig_MURMUR_HASH_2,
			},
		},
		CommonLbConfig: &cluster.Cluster_CommonLbConfig{
			HealthyPanicThreshold:      &_type.Percent{Value: 0},
			IgnoreNewHostsUntilFirstHc: true,
		},
		UpstreamConnectionOptions: &cluster.UpstreamConnectionOptions{
			TcpKeepalive: &core.TcpKeepalive{
				KeepaliveTime:     &wrappers.UInt32Value{Value: 60},
				KeepaliveInterval: &wrappers.UInt32Value{Value: 2},
			},
		},
		CloseConnectionsOnHostHealthFailure: true,
		IgnoreHealthOnHostRemoval:           true,
	}
}
