package tcp

import (
	"time"

	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	tcp_proxy "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/tcp_proxy/v3"
	_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
)

const (
	defaultTTL = time.Second * 30
)

type Option func(*option)

type option struct {
	TTL time.Duration
}

func WithTTL(ttl time.Duration) Option {
	return func(opt *option) {
		if ttl > 0 {
			opt.TTL = ttl
		}
	}
}

func New(name, cluster string, opts ...Option) *listener.Filter {
	opt := new(option)
	for _, f := range opts {
		f(opt)
	}

	if opt.TTL == 0 {
		opt.TTL = defaultTTL
	}

	return &listener.Filter{
		Name: wellknown.TCPProxy,
		ConfigType: &listener.Filter_TypedConfig{
			TypedConfig: func() *anypb.Any {
				config, _ := anypb.New(&tcp_proxy.TcpProxy{
					StatPrefix: name,
					ClusterSpecifier: &tcp_proxy.TcpProxy_Cluster{
						Cluster: cluster,
					},
					IdleTimeout:        durationpb.New(opt.TTL),
					AccessLog:          nil, // TODO
					MaxConnectAttempts: &wrappers.UInt32Value{Value: 2},
					HashPolicy: []*_type.HashPolicy{{
						PolicySpecifier: &_type.HashPolicy_SourceIp_{
							SourceIp: &_type.HashPolicy_SourceIp{},
						},
					}},
				})

				return config
			}(),
		},
	}
}
