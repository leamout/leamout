package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/security/authn"
)

type stubSessionResolver struct {
	token   string
	session authn.Session
}

func (s stubSessionResolver) ResolveSession(_ context.Context, token string) (authn.Session, error) {
	if token != s.token {
		return authn.Session{}, authn.ErrInvalidCredential
	}
	return s.session, nil
}

type stubOrganizationTokenResolver struct {
	token             string
	organizationToken authn.OrganizationToken
}

func (s stubOrganizationTokenResolver) ResolveOrganizationToken(_ context.Context, token string) (authn.OrganizationToken, error) {
	if token != s.token {
		return authn.OrganizationToken{}, authn.ErrInvalidCredential
	}
	return s.organizationToken, nil
}

func TestAuthenticatedOrganizationAllowsSessionWithOrganizationHeader(t *testing.T) {
	organizationID := uuid.New()
	userID := uuid.New()
	authnMiddleware := NewAuthnMiddleware(authn.NewResolver(
		stubSessionResolver{token: "session-token", session: authn.Session{ID: uuid.New(), UserID: userID}},
		stubOrganizationTokenResolver{},
	))
	organizationMiddleware := NewOrganizationMiddleware()

	var gotOrganizationID uuid.UUID
	handler := authnMiddleware.RequireAuthenticated(organizationMiddleware.Require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		gotOrganizationID, ok = OrganizationIDFromContext(r.Context())
		if !ok {
			t.Fatal("expected organization in request context")
		}
		w.WriteHeader(http.StatusNoContent)
	})))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/example", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session-token"})
	req.Header.Set(organizationIDHeader, organizationID.String())
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, res.Code)
	}
	if gotOrganizationID != organizationID {
		t.Fatalf("expected organization %s, got %s", organizationID, gotOrganizationID)
	}
}

func TestAuthenticatedOrganizationAllowsBearerOrganizationToken(t *testing.T) {
	organizationID := uuid.New()
	authnMiddleware := NewAuthnMiddleware(authn.NewResolver(
		stubSessionResolver{},
		stubOrganizationTokenResolver{
			token: "org-token",
			organizationToken: authn.OrganizationToken{
				ID:             uuid.New(),
				OrganizationID: organizationID,
				Scopes:         []string{"organization:read"},
			},
		},
	))
	organizationMiddleware := NewOrganizationMiddleware()

	var gotOrganizationID uuid.UUID
	handler := authnMiddleware.RequireAuthenticated(organizationMiddleware.Require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		gotOrganizationID, ok = OrganizationIDFromContext(r.Context())
		if !ok {
			t.Fatal("expected organization in request context")
		}
		w.WriteHeader(http.StatusNoContent)
	})))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/example", nil)
	req.Header.Set("Authorization", "Bearer org-token")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected %d, got %d", http.StatusNoContent, res.Code)
	}
	if gotOrganizationID != organizationID {
		t.Fatalf("expected organization %s, got %s", organizationID, gotOrganizationID)
	}
}

func TestAuthenticatedOrganizationRequiresOrganizationHeaderForSession(t *testing.T) {
	userID := uuid.New()
	authnMiddleware := NewAuthnMiddleware(authn.NewResolver(
		stubSessionResolver{token: "session-token", session: authn.Session{ID: uuid.New(), UserID: userID}},
		stubOrganizationTokenResolver{},
	))
	organizationMiddleware := NewOrganizationMiddleware()

	handler := authnMiddleware.RequireAuthenticated(organizationMiddleware.Require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})))

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/example", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session-token"})
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, res.Code)
	}
}
