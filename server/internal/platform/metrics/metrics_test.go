package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
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
		"leamout_http_requests_total 1",
		"leamout_http_errors_total 1",
		"leamout_process_uptime_seconds",
		"leamout_process_goroutines",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics response does not contain %q", want)
		}
	}
}
