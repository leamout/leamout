package worker

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWorkerReadinessHandler(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		body   string
	}{
		{name: "ready after subscription", status: http.StatusOK, body: "ready\n"},
		{name: "dependency unavailable", err: errors.New("FreeSWITCH disconnected"), status: http.StatusServiceUnavailable, body: "worker dependencies unavailable\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := workerReadinessHandler(func(context.Context) error { return tt.err })
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/readyz", nil))
			if recorder.Code != tt.status {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.status)
			}
			if got := recorder.Body.String(); got != tt.body {
				t.Fatalf("body = %q, want %q", got, tt.body)
			}
			if strings.Contains(recorder.Body.String(), "FreeSWITCH disconnected") {
				t.Fatal("readiness response leaked dependency details")
			}
		})
	}
}

func TestWorkerReadinessRejectsOtherMethods(t *testing.T) {
	handler := workerReadinessHandler(func(context.Context) error {
		t.Fatal("readiness check must not run for POST")
		return nil
	})
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/readyz", nil))
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusMethodNotAllowed)
	}
}

// Worker metrics remain available separately from readiness, and do not leak
// internal errors or identity-cardinality labels to the HTTP response.
func TestWorkerMetricsHandler(t *testing.T) {
	handler := workerReadinessHandler(
		func(context.Context) error { return errors.New("FreeSWITCH disconnected") },
		func(context.Context) (string, error) {
			return "# TYPE leamout_freeswitch_connected gauge\nleamout_freeswitch_connected 0\n", nil
		},
	)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "leamout_freeswitch_connected 0") {
		t.Fatalf("metrics during outage: status=%d body=%q", recorder.Code, recorder.Body.String())
	}
}
