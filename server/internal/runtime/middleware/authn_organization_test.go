package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/security/authn"
)

type stubMembershipReader struct {
	organizationID uuid.UUID
	userID         uuid.UUID
	role           string
}

func (s stubMembershipReader) GetOrganizationMember(_ context.Context, arg sqlc.GetOrganizationMemberParams) (sqlc.OrganizationMember, error) {
	if arg.OrganizationID != s.organizationID || arg.UserID != s.userID {
		return sqlc.OrganizationMember{}, pgx.ErrNoRows
	}
	return sqlc.OrganizationMember{OrganizationID: s.organizationID, UserID: s.userID, Role: s.role}, nil
}

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
	organizationMiddleware := NewOrganizationMiddleware(stubMembershipReader{
		organizationID: organizationID,
		userID:         userID,
		role:           "member",
	})

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
	organizationMiddleware := NewOrganizationMiddleware(nil)

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
	organizationMiddleware := NewOrganizationMiddleware(nil)

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

func TestAuthenticatedOrganizationRejectsSessionFromAnotherOrganization(t *testing.T) {
	userID := uuid.New()
	requestedOrganizationID := uuid.New()
	authnMiddleware := NewAuthnMiddleware(authn.NewResolver(
		stubSessionResolver{token: "session-token", session: authn.Session{ID: uuid.New(), UserID: userID}},
		stubOrganizationTokenResolver{},
	))
	organizationMiddleware := NewOrganizationMiddleware(stubMembershipReader{
		organizationID: uuid.New(),
		userID:         userID,
		role:           "owner",
	})
	handler := authnMiddleware.RequireAuthenticated(organizationMiddleware.Require(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler should not be called")
	})))
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/numbers", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "session-token"})
	req.Header.Set(organizationIDHeader, requestedOrganizationID.String())
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, res.Code)
	}
}

func TestOrganizationAccessEnforcesSessionRoleAndTokenScope(t *testing.T) {
	organizationID := uuid.New()
	userID := uuid.New()
	tests := []struct {
		name       string
		method     string
		resource   string
		principal  authn.Principal
		role       string
		wantStatus int
	}{
		{name: "member may read", method: http.MethodGet, principal: sessionPrincipal(userID), role: "member", wantStatus: http.StatusNoContent},
		{name: "member may not write", method: http.MethodPost, principal: sessionPrincipal(userID), role: "member", wantStatus: http.StatusForbidden},
		{name: "admin may write", method: http.MethodPost, principal: sessionPrincipal(userID), role: "admin", wantStatus: http.StatusNoContent},
		{name: "token with read scope may read", method: http.MethodGet, principal: tokenPrincipal(organizationID, "numbers:read"), wantStatus: http.StatusNoContent},
		{name: "token with read scope may not write", method: http.MethodPost, principal: tokenPrincipal(organizationID, "numbers:read"), wantStatus: http.StatusForbidden},
		{name: "empty token scopes deny access", method: http.MethodGet, principal: tokenPrincipal(organizationID), wantStatus: http.StatusForbidden},
		{name: "only owners may delete organizations", method: http.MethodDelete, resource: "organization", principal: sessionPrincipal(userID), role: "admin", wantStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewOrganizationMiddleware(stubMembershipReader{organizationID: organizationID, userID: userID, role: tt.role})
			resource := tt.resource
			if resource == "" {
				resource = "numbers"
			}
			handler := middleware.Require(middleware.RequireAccess(resource)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})))
			req := httptest.NewRequestWithContext(context.Background(), tt.method, "/v1/numbers", nil)
			req.Header.Set(organizationIDHeader, organizationID.String())
			req = req.WithContext(authn.WithPrincipal(req.Context(), tt.principal))
			res := httptest.NewRecorder()

			handler.ServeHTTP(res, req)

			if res.Code != tt.wantStatus {
				t.Fatalf("expected %d, got %d", tt.wantStatus, res.Code)
			}
		})
	}
}

func sessionPrincipal(userID uuid.UUID) authn.Principal {
	return authn.Principal{
		Subject:    authn.Subject{ID: userID, Type: authn.SubjectUser},
		Credential: authn.Credential{ID: uuid.New(), Type: authn.CredentialSession},
	}
}

func tokenPrincipal(organizationID uuid.UUID, scopes ...string) authn.Principal {
	credentialID := uuid.New()
	return authn.Principal{
		Subject:        authn.Subject{ID: credentialID, Type: authn.SubjectOrganizationToken},
		Credential:     authn.Credential{ID: credentialID, Type: authn.CredentialOrganizationToken},
		OrganizationID: organizationID,
		Scopes:         scopes,
	}
}
