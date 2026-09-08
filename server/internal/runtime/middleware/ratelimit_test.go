package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func TestRateLimitAllowsRequestsWithinOrganizationBudget(t *testing.T) {
	store := memory.NewStore()
	middleware, err := NewRateLimitMiddleware(store)
	if err != nil {
		t.Fatal(err)
	}

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/calls", nil)
	req = req.WithContext(withOrganizationContext(req.Context(), organizationContext{ID: uuid.New()}))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, res.Code)
	}
	if res.Header().Get("X-RateLimit-Limit") != "1000" {
		t.Fatalf("expected rate limit header, got %q", res.Header().Get("X-RateLimit-Limit"))
	}
}

func TestRateLimitRejectsOrganizationAfterBudgetIsExhausted(t *testing.T) {
	store := memory.NewStore()
	middleware, err := NewRateLimitMiddleware(store)
	if err != nil {
		t.Fatal(err)
	}

	organizationID := uuid.New()
	for range 1000 {
		if _, err := middleware.limiter.Get(context.Background(), "http:organization:"+organizationID.String()); err != nil {
			t.Fatal(err)
		}
	}

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/calls", nil)
	req = req.WithContext(withOrganizationContext(req.Context(), organizationContext{ID: organizationID}))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusTooManyRequests {
		t.Fatalf("expected %d, got %d", http.StatusTooManyRequests, res.Code)
	}
	if res.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}

func TestRateLimitKeepsOrganizationBudgetsIndependent(t *testing.T) {
	store := memory.NewStore()
	middleware, err := NewRateLimitMiddleware(store)
	if err != nil {
		t.Fatal(err)
	}

	exhaustedOrganizationID := uuid.New()
	for range 1000 {
		if _, err := middleware.limiter.Get(context.Background(), "http:organization:"+exhaustedOrganizationID.String()); err != nil {
			t.Fatal(err)
		}
	}

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/calls", nil)
	req = req.WithContext(withOrganizationContext(req.Context(), organizationContext{ID: uuid.New()}))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected independent organization to receive %d, got %d", http.StatusNoContent, res.Code)
	}
}

func TestRateLimitFailsClosedWhenStoreIsUnavailable(t *testing.T) {
	middleware, err := NewRateLimitMiddleware(failingRateLimitStore{})
	if err != nil {
		t.Fatal(err)
	}

	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/calls", nil)
	req = req.WithContext(withOrganizationContext(req.Context(), organizationContext{ID: uuid.New()}))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, res.Code)
	}
}

type failingRateLimitStore struct{}

func (failingRateLimitStore) Get(context.Context, string, limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, errors.New("store unavailable")
}

func (failingRateLimitStore) Peek(context.Context, string, limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, errors.New("store unavailable")
}

func (failingRateLimitStore) Reset(context.Context, string, limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, errors.New("store unavailable")
}

func (failingRateLimitStore) Increment(context.Context, string, int64, limiter.Rate) (limiter.Context, error) {
	return limiter.Context{}, errors.New("store unavailable")
}
