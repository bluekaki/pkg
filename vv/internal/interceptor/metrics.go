package interceptor

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "byepichi"
	subsystem = "vv"
)

func init() {
	prometheus.MustRegister(MetricsRequestCost)
	prometheus.MustRegister(MetricsError)
}

// all metrics used by WithPrometheus & WithPrometheusPush

// MetricsRequestCost metrics for ok request cost
var MetricsRequestCost = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "requestcost",
	Help:      "[ok] request(s) cost seconds",
	Buckets:   []float64{0.1, 0.3, 0.5, 0.7, 0.9, 1.1},
}, []string{"method"})

// MetricsError metrics for alertmanager
var MetricsError = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "error",
	Help:      "error(s) alert",
	Buckets:   []float64{0.1, 0.3, 0.5, 0.7, 0.9, 1.1},
}, []string{"method", "code", "message", "journal_id"})
