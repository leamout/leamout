package subscriptions

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/internal/security/authn"
)

func TestSubscriptionRoutesRejectInvalidIDs(t *testing.T) {
	router := chi.NewRouter()
	organizationID := uuid.New()
	auth := func(next http.Handler) http.Handler {
		principal := authn.Principal{
			Subject:        authn.Subject{ID: uuid.New(), Type: authn.SubjectOrganizationToken},
			Credential:     authn.Credential{ID: uuid.New(), Type: authn.CredentialOrganizationToken},
			OrganizationID: organizationID,
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(authn.WithPrincipal(r.Context(), principal)))
		})
	}
	organizationAuth := middleware.NewOrganizationMiddleware().Require
	RegisterRoutes(router, NewHandler(&Service{}), func(next http.Handler) http.Handler {
		return auth(organizationAuth(next))
	})

	for _, request := range []*http.Request{
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/subscriptions/not-a-uuid", nil),
		httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/subscriptions/not-a-uuid/cancel", nil),
		httptest.NewRequestWithContext(t.Context(), http.MethodPatch, "/subscriptions/not-a-uuid", bytes.NewBufferString(`{"price_id":"bad"}`)),
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Errorf("%s %s: expected status 400, got %d", request.Method, request.URL.Path, response.Code)
		}
	}
}

func TestSubscriptionHandlerRequiresOrganizationContext(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/subscriptions", nil)
	NewHandler(&Service{}).List(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestSubscriptionResponse(t *testing.T) {
	subscription := Subscription{
		ID: uuid.New(), OrganizationID: uuid.New(), PlanID: uuid.New(), Status: StatusActive,
		StartsAt: time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC),
	}
	encoded, err := json.Marshal(newSubscriptionResponse(subscription))
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	var response map[string]any
	if err := json.Unmarshal(encoded, &response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response["status"] != string(StatusActive) {
		t.Fatalf("expected active status, got %q", response["status"])
	}
	if _, exposed := response["provider_subscription_id"]; exposed {
		t.Fatalf("response exposes provider reconciliation metadata: %s", encoded)
	}
}
