package worker

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stubDependency struct{ err error }

func (s stubDependency) Ping(context.Context) error { return s.err }

type stubMedia struct{ err error }

func (s stubMedia) HealthCheck(context.Context) error { return s.err }

func TestHealthHandlerReadinessIncludesDependenciesAndComponents(t *testing.T) {
	state := newHealthState("outbox-publisher", "webhook-consumer")
	handler := healthHandler(stubDependency{}, stubDependency{}, stubDependency{}, stubMedia{}, state)

	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected starting components to make readiness %d, got %d", http.StatusServiceUnavailable, res.Code)
	}

	state.setRunning("outbox-publisher")
	state.setRunning("webhook-consumer")
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if res.Code != http.StatusNoContent {
		t.Fatalf("expected ready status %d, got %d: %s", http.StatusNoContent, res.Code, res.Body.String())
	}

	state.setStopped("webhook-consumer", errors.New("subscription lost"))
	res = httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if res.Code != http.StatusServiceUnavailable || !strings.Contains(res.Body.String(), "subscription lost") {
		t.Fatalf("expected failed component in readiness response, got %d: %s", res.Code, res.Body.String())
	}
}

func TestHealthHandlerRejectsUnavailableDependency(t *testing.T) {
	state := newHealthState("outbox-publisher")
	state.setRunning("outbox-publisher")
	handler := healthHandler(stubDependency{}, stubDependency{err: errors.New("nats down")}, stubDependency{}, stubMedia{}, state)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	if res.Code != http.StatusServiceUnavailable || !strings.Contains(res.Body.String(), `"nats":"unavailable"`) {
		t.Fatalf("expected NATS readiness failure, got %d: %s", res.Code, res.Body.String())
	}
}

func TestHealthHandlerExposesComponentMetrics(t *testing.T) {
	state := newHealthState("webhook-consumer", "outbox-publisher")
	state.setRunning("outbox-publisher")
	handler := healthHandler(stubDependency{}, stubDependency{}, stubDependency{}, stubMedia{}, state)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	body := res.Body.String()
	if !strings.Contains(body, `leamout_worker_component_up{component="outbox-publisher"} 1`) ||
		!strings.Contains(body, `leamout_worker_component_up{component="webhook-consumer"} 0`) {
		t.Fatalf("unexpected metrics response: %s", body)
	}
}
