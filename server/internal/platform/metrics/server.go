package metrics

import (
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
	"time"
)

// Registry contains process metrics that are safe to update concurrently.
//
// The registry intentionally stays dependency-free. It provides the small
// platform-level surface needed by the runtime while keeping the transport
// format behind Handler.
type Registry struct {
	requests atomic.Uint64
	errors   atomic.Uint64
}

// IncRequests increments the total request counter.
func (r *Registry) IncRequests() {
	r.requests.Add(1)
}

// IncErrors increments the total error counter.
func (r *Registry) IncErrors() {
	r.errors.Add(1)
}

// Requests returns the total number of requests recorded by the registry.
func (r *Registry) Requests() uint64 {
	return r.requests.Load()
}

// Errors returns the total number of errors recorded by the registry.
func (r *Registry) Errors() uint64 {
	return r.errors.Load()
}

// Handler exposes the registry using Prometheus' text exposition format.
//
// The output is intentionally small for now. Additional application and
// telecom metrics can be added to Registry without coupling callers to an
// external metrics client.
func Handler(registry *Registry) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")

		_, _ = fmt.Fprintf(w,
			"# HELP leamout_http_requests_total Total number of HTTP requests recorded.\n"+
				"# TYPE leamout_http_requests_total counter\n"+
				"leamout_http_requests_total %d\n"+
				"# HELP leamout_http_errors_total Total number of HTTP errors recorded.\n"+
				"# TYPE leamout_http_errors_total counter\n"+
				"leamout_http_errors_total %d\n"+
				"# HELP leamout_process_uptime_seconds Process uptime in seconds.\n"+
				"# TYPE leamout_process_uptime_seconds gauge\n"+
				"leamout_process_uptime_seconds %f\n"+
				"# HELP leamout_process_goroutines Current number of goroutines.\n"+
				"# TYPE leamout_process_goroutines gauge\n"+
				"leamout_process_goroutines %d\n",
			registry.Requests(),
			registry.Errors(),
			time.Since(startTime).Seconds(),
			runtime.NumGoroutine(),
		)
	})
}

var startTime = time.Now()
