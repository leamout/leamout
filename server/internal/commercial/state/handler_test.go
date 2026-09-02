package state

import (
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

func TestCommercialStateRouteValidatesOrganizationID(t *testing.T) {
	organizationID := uuid.New()
	router := commercialStateTestRouter(organizationID)

	for _, test := range []struct {
		name   string
		path   string
		status int
	}{
		{name: "invalid ID", path: "/organizations/not-a-uuid/commercial-state", status: http.StatusBadRequest},
		{name: "mismatched organization", path: "/organizations/" + uuid.NewString() + "/commercial-state", status: http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, test.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("expected status %d, got %d", test.status, response.Code)
			}
		})
	}
}

func TestOrganizationStateResponse(t *testing.T) {
	organizationID := uuid.New()
	subscriptionID := uuid.New()
	planID := uuid.New()
	effectiveAt := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	response := newOrganizationStateResponse(OrganizationState{
		OrganizationID: organizationID,
		Standing:       StandingActive,
		SubscriptionID: &subscriptionID,
		PlanID:         &planID,
		Features:       map[string]bool{"voice": true},
		Limits:         map[string]int64{"max.deployments": 2},
		EffectiveAt:    effectiveAt,
	})

	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if decoded["organization_id"] != organizationID.String() {
		t.Fatalf("unexpected organization_id: %v", decoded["organization_id"])
	}
	if decoded["standing"] != string(StandingActive) {
		t.Fatalf("unexpected standing: %v", decoded["standing"])
	}
}

func commercialStateTestRouter(organizationID uuid.UUID) http.Handler {
	router := chi.NewRouter()
	organizationAuth := middleware.NewOrganizationMiddleware().Require
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
	RegisterRoutes(router, NewHandler(&Service{}), func(next http.Handler) http.Handler {
		return auth(organizationAuth(next))
	})
	return router
}
