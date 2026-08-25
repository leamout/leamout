package metrics

import "github.com/prometheus/client_golang/prometheus"

const namespace = "leamout"

// Metrics owns the process-wide application metrics registry.
type Metrics struct {
	Registry *prometheus.Registry

	HTTPRequests *prometheus.CounterVec
	HTTPDuration *prometheus.HistogramVec
}

// New creates an isolated Prometheus registry for the Leamout process.
func New() *Metrics {
	registry := prometheus.NewRegistry()

	metrics := &Metrics{
		Registry: registry,
		HTTPRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "requests_total",
				Help:      "Total number of HTTP requests handled by the server.",
			},
			[]string{"method", "route", "status"},
		),
		HTTPDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Subsystem: "http",
				Name:      "request_duration_seconds",
				Help:      "HTTP request duration in seconds.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"method", "route"},
		),
	}

	registry.MustRegister(
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
		metrics.HTTPRequests,
		metrics.HTTPDuration,
	)

	return metrics
}
