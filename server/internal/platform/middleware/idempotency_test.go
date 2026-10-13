package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIdempotencyRequiredRejectsMissingKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
	}{
		{name: "missing"},
		{name: "blank", key: "   "},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			invoked := false
			handler := (&IdempotencyMiddleware{}).HandleRequired(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				invoked = true
				w.WriteHeader(http.StatusCreated)
			}))
			req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/billing/checkout", nil)
			if tc.key != "" {
				req.Header.Set("Idempotency-Key", tc.key)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
			if invoked {
				t.Fatal("checkout handler must not run without an idempotency key")
			}
		})
	}
}

func TestIdempotencyOptionalStillPassesMissingKey(t *testing.T) {
	t.Parallel()

	invoked := false
	handler := (&IdempotencyMiddleware{}).Handle(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		invoked = true
		w.WriteHeader(http.StatusAccepted)
	}))
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/calls", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !invoked || rec.Code != http.StatusAccepted {
		t.Fatalf("optional-key behavior changed: invoked=%v, status=%d", invoked, rec.Code)
	}
}

func TestIdempotencyRequiredDoesNotBlockReads(t *testing.T) {
	t.Parallel()

	invoked := false
	handler := (&IdempotencyMiddleware{}).HandleRequired(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		invoked = true
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/billing", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !invoked || rec.Code != http.StatusOK {
		t.Fatalf("GET blocked: invoked=%v, status=%d", invoked, rec.Code)
	}
}
