package interceptor

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "bluekaki"
	subsystem = "vv"
)

func init() {
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestDuration)
}

var requestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "request_total",
}, []string{"class", "method", "success"})

var requestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "request_duration",
	Buckets:   []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1},
}, []string{"class", "method", "success"})
