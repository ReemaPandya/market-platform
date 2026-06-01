package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var HttpRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total HTTP requests",
	},
	[]string{"path", "method"},
)

func RegisterMetrics() {

	prometheus.MustRegister(HttpRequests)
}
