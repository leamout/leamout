package backoffice

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/security/authn"
)

type stubUserLookup struct {
	user sqlc.User
	err  error
}

func (s stubUserLookup) GetUserByID(context.Context, uuid.UUID) (sqlc.User, error) {
	return s.user, s.err
}

func TestRequirePlatformAdminAllowsAdmin(t *testing.T) {
	userID := uuid.New()
	middleware := requirePlatformAdmin(stubUserLookup{user: sqlc.User{
		ID:              userID,
		IsPlatformAdmin: true,
	}})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	recorder := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	principal := sessionPrincipal(userID)
	handler.ServeHTTP(recorder, r.WithContext(authn.WithPrincipal(r.Context(), principal)))

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestRequirePlatformAdminRejectsNonAdmin(t *testing.T) {
	userID := uuid.New()
	middleware := requirePlatformAdmin(stubUserLookup{user: sqlc.User{ID: userID}})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	recorder := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	principal := sessionPrincipal(userID)
	handler.ServeHTTP(recorder, r.WithContext(authn.WithPrincipal(r.Context(), principal)))

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestRequirePlatformAdminRequiresSessionPrincipal(t *testing.T) {
	middleware := requirePlatformAdmin(stubUserLookup{})
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	recorder := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func sessionPrincipal(userID uuid.UUID) authn.Principal {
	return authn.Principal{
		Subject: authn.Subject{
			ID:   userID,
			Type: authn.SubjectUser,
		},
		Credential: authn.Credential{
			ID:   uuid.New(),
			Type: authn.CredentialSession,
		},
		Assurance: authn.AssurancePassword,
	}
}
