package interceptor

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "bluekaki"
	subsystem = "vv"
)

func init() {
	prometheus.MustRegister(requestsCounter)
}

var requestsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Namespace: namespace,
	Subsystem: subsystem,
	Name:      "requests_total",
}, []string{"class", "method", "success"})
