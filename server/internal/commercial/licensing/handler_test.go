package licensing

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

func TestLicenseRoutesRejectInvalidLicenseIDs(t *testing.T) {
	organizationID := uuid.New()
	router := commercialLicenseTestRouter(organizationID)

	requests := []*http.Request{
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/licenses/not-a-uuid", nil),
		httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/licenses/not-a-uuid/deployments", nil),
		httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/licenses/not-a-uuid/deployments", bytes.NewBufferString(`{}`)),
		httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/licenses/not-a-uuid/deployments/node-01/heartbeat", nil),
		httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/licenses/not-a-uuid/deployments/node-01", nil),
	}
	for _, request := range requests {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Errorf("%s %s: expected status 400, got %d", request.Method, request.URL.Path, response.Code)
		}
	}
}

func TestLicenseHandlerRequiresOrganizationContext(t *testing.T) {
	response := httptest.NewRecorder()
	request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/licenses", nil)
	NewHandler(&Service{}).List(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestLicenseRoutesDoNotExposeCustomerCreation(t *testing.T) {
	organizationID := uuid.New()
	router := commercialLicenseTestRouter(organizationID)
	request := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/licenses", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", response.Code)
	}
}

func TestLicenseResponsesHideSigningMetadata(t *testing.T) {
	now := time.Date(2026, 9, 2, 12, 0, 0, 0, time.UTC)
	keyID := "private-authority-key-reference"
	values := []any{
		newLicenseResponse(License{
			ID: uuid.New(), OrganizationID: uuid.New(), Status: StatusPending,
			MaxDeployments: 2, SigningKeyID: &keyID, IssuedAt: now,
		}),
		newDeploymentResponse(Deployment{
			ID: uuid.New(), LicenseID: uuid.New(), DeploymentID: "node-01",
			Status: DeploymentStatusActive, ActivatedAt: now,
		}),
	}

	for _, value := range values {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("marshal response: %v", err)
		}
		var response map[string]any
		if err := json.Unmarshal(encoded, &response); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if response["id"] == "" {
			t.Fatalf("expected response ID, got %s", encoded)
		}
		if _, exposed := response["signing_key_id"]; exposed {
			t.Fatalf("response exposes signing authority metadata: %s", encoded)
		}
	}
}

func commercialLicenseTestRouter(organizationID uuid.UUID) http.Handler {
	router := chi.NewRouter()
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
	}, func(next http.Handler) http.Handler { return next })
	return router
}
