package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestRegistry(t *testing.T) {
	registry := New()

	registry.IncRequests()
	registry.IncRequests()
	registry.IncErrors()

	if got := registry.Requests(); got != 2 {
		t.Fatalf("Requests() = %d, want 2", got)
	}

	if got := registry.Errors(); got != 1 {
		t.Fatalf("Errors() = %d, want 1", got)
	}
}

func TestRegistryConcurrent(t *testing.T) {
	registry := New()

	const (
		workers = 16
		updates = 1_000
	)

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < updates; j++ {
				registry.IncRequests()
				registry.IncErrors()
			}
		}()
	}

	wg.Wait()

	want := uint64(workers * updates)
	if got := registry.Requests(); got != want {
		t.Fatalf("Requests() = %d, want %d", got, want)
	}

	if got := registry.Errors(); got != want {
		t.Fatalf("Errors() = %d, want %d", got, want)
	}
}

func TestHandler(t *testing.T) {
	registry := New()
	registry.IncRequests()
	registry.IncErrors()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	recorder := httptest.NewRecorder()

	Handler(registry).ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	if got := recorder.Header().Get("Content-Type"); got != "text/plain; version=0.0.4" {
		t.Fatalf("Content-Type = %q, want Prometheus text format", got)
	}

	body := recorder.Body.String()
	for _, want := range []string{
		"# HELP leamout_http_requests_total Total number of HTTP requests recorded.",
		"# TYPE leamout_http_requests_total counter",
		"leamout_http_requests_total 1",
		"# HELP leamout_http_errors_total Total number of HTTP errors recorded.",
		"# TYPE leamout_http_errors_total counter",
		"leamout_http_errors_total 1",
		"# HELP leamout_process_uptime_seconds Process uptime in seconds.",
		"# TYPE leamout_process_uptime_seconds gauge",
		"leamout_process_uptime_seconds",
		"# HELP leamout_process_goroutines Current number of goroutines.",
		"# TYPE leamout_process_goroutines gauge",
		"leamout_process_goroutines",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics response does not contain %q", want)
		}
	}
}
