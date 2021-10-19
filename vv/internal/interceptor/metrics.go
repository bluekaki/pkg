package interceptor

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "bluekaki"
	subsystem = "vv"
)

func init() {
	prometheus.MustRegister(httpRequestSuccessCounter)
	prometheus.MustRegister(grpcRequestSuccessCounter)
	prometheus.MustRegister(httpRequestErrorCounter)
	prometheus.MustRegister(grpcRequestErrorCounter)
	prometheus.MustRegister(httpRequestSuccessDurationHistogram)
	prometheus.MustRegister(grpcRequestSuccessDurationHistogram)
	prometheus.MustRegister(httpRequestErrorDurationHistogram)
	prometheus.MustRegister(grpcRequestErrorDurationHistogram)
}

var httpRequestSuccessCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "http_request_success_total",
}, []string{"method"})

var grpcRequestSuccessCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "grpc_request_success_total",
}, []string{"method"})

var httpRequestErrorCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "http_request_error_total",
}, []string{"method", "code"})

var grpcRequestErrorCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "grpc_request_error_total",
}, []string{"method", "code"})

var httpRequestSuccessDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "http_request_success_duration_seconds",
	Buckets:   []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1},
}, []string{"method"})

var grpcRequestSuccessDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "grpc_request_success_duration_seconds",
	Buckets:   []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1},
}, []string{"method"})

var httpRequestErrorDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "http_request_error_duration_seconds",
	Buckets:   []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1},
}, []string{"method", "code"})

var grpcRequestErrorDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "grpc_request_error_duration_seconds",
	Buckets:   []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1},
}, []string{"method", "code"})
