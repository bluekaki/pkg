package router

import (
	"net/http"
	"time"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	local_ratelimit "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/local_ratelimit/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	_type "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/golang/protobuf/ptypes/wrappers"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	defaultTTL       = time.Second * 30
	defaultRateLimit = 1000
)

var cors = &route.CorsPolicy{
	AllowOriginStringMatch: []*matcher.StringMatcher{{
		MatchPattern: &matcher.StringMatcher_SafeRegex{
			SafeRegex: &matcher.RegexMatcher{
				EngineType: &matcher.RegexMatcher_GoogleRe2{
					GoogleRe2: &matcher.RegexMatcher_GoogleRE2{},
				},
				Regex: `\*`,
			},
		},
	}},
	AllowMethods: "OPTIONS, GET, POST, PUT, PATCH, DELETE",
	AllowHeaders: "*",
	AllowCredentials: &wrappers.BoolValue{
		Value: true,
	},
}

var retryPolicy = &route.RetryPolicy{
	RetryOn:                       "gateway-error,reset,envoy-ratelimited,retriable-status-codes,retriable-headers,cancelled,deadline-exceeded,resource-exhausted,unavailable",
	NumRetries:                    &wrappers.UInt32Value{Value: 5},
	HostSelectionRetryMaxAttempts: 3,
	RetriableStatusCodes: []uint32{
		http.StatusRequestTimeout,
		http.StatusLocked,
		http.StatusTooEarly,
	},
	RetryBackOff: &route.RetryPolicy_RetryBackOff{
		BaseInterval: durationpb.New(time.Second),
	},
}

type Option func(*option)

type option struct {
	GRPC                 bool
	WebSocket            bool
	TTL                  time.Duration
	RateLimit            uint32
	CORS                 *route.CorsPolicy
	Authority            string
	Headers              []*route.HeaderMatcher
	DisablePrefixRewrite bool
}

func WithGRPC() Option {
	return func(opt *option) {
		opt.GRPC = true
	}
}

func WithWebSocket() Option {
	return func(opt *option) {
		opt.WebSocket = true
	}
}

func WithTTL(ttl time.Duration) Option {
	return func(opt *option) {
		if ttl > 0 {
			opt.TTL = ttl
		}
	}
}

func WithCORS() Option {
	return func(opt *option) {
		opt.CORS = cors
	}
}

func WithRateLimit(ratelimit uint32) Option {
	return func(opt *option) {
		if ratelimit > 0 {
			opt.RateLimit = ratelimit
		}
	}
}

func WithAuthority(authority string) Option {
	return func(opt *option) {
		opt.Authority = authority
		opt.Headers = append(opt.Headers, &route.HeaderMatcher{
			Name: ":authority",
			HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
				StringMatch: &matcher.StringMatcher{
					MatchPattern: &matcher.StringMatcher_Exact{
						Exact: authority,
					},
				},
			},
		})
	}
}

func WithHeader(key, value string, exactMatch bool) Option {
	return func(opt *option) {
		if exactMatch {
			opt.Headers = append(opt.Headers, &route.HeaderMatcher{
				Name: key,
				HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
					StringMatch: &matcher.StringMatcher{
						MatchPattern: &matcher.StringMatcher_Exact{
							Exact: value,
						},
					},
				},
			})

		} else {
			opt.Headers = append(opt.Headers, &route.HeaderMatcher{
				Name: key,
				HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
					StringMatch: &matcher.StringMatcher{
						MatchPattern: &matcher.StringMatcher_Contains{
							Contains: value,
						},
					},
				},
			})
		}
	}
}

func WithDisablePrefixRewrite() Option {
	return func(opt *option) {
		opt.DisablePrefixRewrite = true
	}
}

func New(routes ...*route.Route) *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name: "http_grpc_rds_config",
		VirtualHosts: []*route.VirtualHost{
			{
				Name:    "*",
				Domains: []string{"*"},
				Routes:  routes,
			},
		},
	}
}

func NewHTTPRoute(name, pathPrefix, cluster string, opts ...Option) *route.Route {
	opt := new(option)
	for _, f := range opts {
		f(opt)
	}

	if opt.TTL == 0 {
		opt.TTL = defaultTTL
	}

	if opt.RateLimit == 0 {
		opt.RateLimit = defaultRateLimit
	}

	if opt.RateLimit = opt.RateLimit / 5; opt.RateLimit == 0 { // per 200ms fill tokens
		opt.RateLimit = 1
	}

	if len(opt.Headers) == 0 {
		panic(name + ": header matcher required")
	}

	return &route.Route{
		Match: &route.RouteMatch{
			PathSpecifier: &route.RouteMatch_Prefix{
				Prefix: pathPrefix,
			},
			Headers: opt.Headers,
			Grpc: func() *route.RouteMatch_GrpcRouteMatchOptions {
				if !opt.GRPC {
					return nil
				}

				return &route.RouteMatch_GrpcRouteMatchOptions{}
			}(),
		},
		Action: &route.Route_Route{
			Route: &route.RouteAction{
				ClusterSpecifier: &route.RouteAction_Cluster{
					Cluster: cluster,
				},
				PrefixRewrite: func() (prefix string) {
					if opt.DisablePrefixRewrite {
						return pathPrefix
					}

					if !opt.GRPC && pathPrefix != "/" {
						return "/"
					}
					return
				}(),
				HostRewriteSpecifier: &route.RouteAction_AutoHostRewrite{
					AutoHostRewrite: &wrappers.BoolValue{Value: true},
				},
				Timeout:               durationpb.New(opt.TTL * 2), // 504 Gateway Timeout
				IdleTimeout:           durationpb.New(opt.TTL),     // 408 Request Timeout
				RetryPolicy:           retryPolicy,
				RequestMirrorPolicies: nil, // TODO ...
				HashPolicy: []*route.RouteAction_HashPolicy{{
					PolicySpecifier: &route.RouteAction_HashPolicy_Header_{
						Header: &route.RouteAction_HashPolicy_Header{
							HeaderName: "journal-id",
						},
					},
					Terminal: true,
				}},
				Cors: opt.CORS,
				UpgradeConfigs: func() []*route.RouteAction_UpgradeConfig {
					if !opt.WebSocket {
						return nil
					}

					return []*route.RouteAction_UpgradeConfig{
						{UpgradeType: "websocket"},
					}
				}(),
				InternalRedirectPolicy: nil, // TODO ...
				MaxStreamDuration:      nil, // TODO ...
			},
		},
		TypedPerFilterConfig: map[string]*anypb.Any{
			"envoy.filters.http.local_ratelimit": func() *anypb.Any {
				config, _ := anypb.New(&local_ratelimit.LocalRateLimit{
					StatPrefix: name + "_local_ratelimit",
					TokenBucket: &_type.TokenBucket{
						MaxTokens:     opt.RateLimit,
						TokensPerFill: wrapperspb.UInt32(opt.RateLimit),
						FillInterval:  durationpb.New(time.Millisecond * 200),
					},
					FilterEnabled: &core.RuntimeFractionalPercent{
						DefaultValue: &_type.FractionalPercent{
							Numerator:   100,
							Denominator: _type.FractionalPercent_HUNDRED,
						},
					},
					FilterEnforced: &core.RuntimeFractionalPercent{
						DefaultValue: &_type.FractionalPercent{
							Numerator:   100,
							Denominator: _type.FractionalPercent_HUNDRED,
						},
					},
				})

				return config
			}(),
		},
		ResponseHeadersToAdd: func() (headers []*core.HeaderValueOption) {
			headers = append(headers, &core.HeaderValueOption{
				Header: &core.HeaderValue{Key: "x-forwarded-host", Value: opt.Authority},
				Append: &wrappers.BoolValue{Value: false},
			})

			headers = append(headers, &core.HeaderValueOption{
				Header: &core.HeaderValue{Key: "x-forwarded-prefix", Value: pathPrefix},
				Append: &wrappers.BoolValue{Value: false},
			})

			return
		}(),
		Tracing: nil, // TODO ...
	}
}
