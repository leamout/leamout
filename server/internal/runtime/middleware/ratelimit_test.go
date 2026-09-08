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

	"github.com/leamout/leamout/internal/security/authn"
)

func TestRateLimitSeparatesReadAndWriteBudgets(t *testing.T) {
	middleware, err := NewRateLimitMiddleware(memory.NewStore())
	if err != nil {
		t.Fatal(err)
	}
	organizationID := uuid.New()
	credentialID := uuid.New()
	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for range 300 {
		if code := serveRateLimited(handler, http.MethodPost, organizationID, credentialID).Code; code != http.StatusNoContent {
			t.Fatalf("expected write within budget to receive %d, got %d", http.StatusNoContent, code)
		}
	}
	if code := serveRateLimited(handler, http.MethodPost, organizationID, credentialID).Code; code != http.StatusTooManyRequests {
		t.Fatalf("expected exhausted write budget to receive %d, got %d", http.StatusTooManyRequests, code)
	}
	if code := serveRateLimited(handler, http.MethodGet, organizationID, credentialID).Code; code != http.StatusNoContent {
		t.Fatalf("expected independent read budget to receive %d, got %d", http.StatusNoContent, code)
	}
}

func TestRateLimitKeepsCredentialBudgetsIndependent(t *testing.T) {
	middleware, err := NewRateLimitMiddleware(memory.NewStore())
	if err != nil {
		t.Fatal(err)
	}
	organizationID := uuid.New()
	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	credentialID := uuid.New()
	for range 600 {
		serveRateLimited(handler, http.MethodGet, organizationID, credentialID)
	}
	if code := serveRateLimited(handler, http.MethodGet, organizationID, credentialID).Code; code != http.StatusTooManyRequests {
		t.Fatalf("expected exhausted credential to receive %d, got %d", http.StatusTooManyRequests, code)
	}

	if code := serveRateLimited(handler, http.MethodGet, organizationID, uuid.New()).Code; code != http.StatusNoContent {
		t.Fatalf("expected independent credential to receive %d, got %d", http.StatusNoContent, code)
	}
}

func TestRateLimitEnforcesSharedOrganizationBudget(t *testing.T) {
	middleware, err := NewRateLimitMiddleware(memory.NewStore())
	if err != nil {
		t.Fatal(err)
	}
	organizationID := uuid.New()
	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, credentialID := range []uuid.UUID{uuid.New(), uuid.New()} {
		for range 600 {
			if code := serveRateLimited(handler, http.MethodGet, organizationID, credentialID).Code; code != http.StatusNoContent {
				t.Fatalf("expected request within shared budget to receive %d, got %d", http.StatusNoContent, code)
			}
		}
	}

	res := serveRateLimited(handler, http.MethodGet, organizationID, uuid.New())
	if res.Code != http.StatusTooManyRequests {
		t.Fatalf("expected exhausted organization to receive %d, got %d", http.StatusTooManyRequests, res.Code)
	}
	if res.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Fatalf("expected zero remaining organization requests, got %q", res.Header().Get("X-RateLimit-Remaining"))
	}
	if res.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
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

	res := serveRateLimited(handler, http.MethodPost, uuid.New(), uuid.New())
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected %d, got %d", http.StatusServiceUnavailable, res.Code)
	}
}

func serveRateLimited(handler http.Handler, method string, organizationID, credentialID uuid.UUID) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, "/v1/example", nil)
	req = req.WithContext(authn.WithPrincipal(req.Context(), authn.Principal{
		Subject:    authn.Subject{ID: uuid.New(), Type: authn.SubjectUser},
		Credential: authn.Credential{ID: credentialID, Type: authn.CredentialOrganizationToken},
	}))
	req = req.WithContext(withOrganizationContext(req.Context(), organizationContext{ID: organizationID}))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
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
