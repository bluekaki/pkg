package server

import (
	"net/http"
	"time"

	"github.com/byepichi/pkg/vv/internal/interceptor"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.uber.org/zap"
)

// WithPrometheus prometheus metrics exposes on http://addr/metrics
func WithPrometheus(addr string) Option {
	return func(opt *option) {
		opt.prometheusHandler = func(logger *zap.Logger) {
			http.Handle("/metrics", promhttp.Handler())
			go func() {
				if err := http.ListenAndServe(addr, nil); err != nil {
					panic(err)
				}
			}()
		}
	}
}

// WithPrometheusPush  push prometheus metrics to the Pushgateway
func WithPrometheusPush(gateway string) Option {
	return func(opt *option) {
		opt.prometheusHandler = func(logger *zap.Logger) {
			go func() {
				pusher := push.New(gateway, "bluekaiki_vv_metrics").
					Collector(interceptor.MetricsRequestCost).
					Collector(interceptor.MetricsError)

				for range time.NewTicker(time.Second * 5).C {
					if err := pusher.Add(); err != nil {
						logger.Error("post metrics to prometheus pushgateway err", zap.Error(err))
					}
				}
			}()
		}
	}
}
