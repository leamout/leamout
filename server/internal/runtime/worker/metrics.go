package worker

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

func workerMetricsHandler() http.Handler {
	started := time.Now()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = fmt.Fprintf(w, "# TYPE leamout_worker_running gauge\nleamout_worker_running 1\n# TYPE leamout_worker_process_goroutines gauge\nleamout_worker_process_goroutines %d\n# TYPE leamout_worker_process_uptime_seconds gauge\nleamout_worker_process_uptime_seconds %f\n", runtime.NumGoroutine(), time.Since(started).Seconds())
	})
	return mux
}
