package backoffice

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	backofficeauth "github.com/leamout/leamout/internal/backoffice/auth"
	"github.com/leamout/leamout/internal/security/authn"
)

type stubAuthenticator struct {
	principal authn.Principal
	err       error
}

func (s stubAuthenticator) Authenticate(context.Context, string) (authn.Principal, error) {
	return s.principal, s.err
}

func authenticatedPrincipal() authn.Principal {
	return authn.Principal{
		Subject: authn.Subject{
			ID:   uuid.New(),
			Type: authn.SubjectUser,
		},
		Credential: authn.Credential{
			ID:   uuid.New(),
			Type: authn.CredentialSession,
		},
	}
}

func testServer(authentication authenticator) *Server {
	return newServer(nil, newModules(nil), nil, authentication)
}

func authenticatedRequest(t *testing.T, method, path string) *http.Request {
	t.Helper()
	r := httptest.NewRequestWithContext(t.Context(), method, path, nil)
	r.AddCookie(&http.Cookie{
		Name:  backofficeauth.CookieName,
		Value: "test-backoffice-session",
	})
	return r
}

func TestHealth(t *testing.T) {
	srv := testServer(nil)
	recorder := httptest.NewRecorder()
	srv.Router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestProtectedRouteRequiresAuthentication(t *testing.T) {
	srv := testServer(nil)
	recorder := httptest.NewRecorder()
	srv.Router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("dashboard status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if location := recorder.Header().Get("Location"); location != "/login" {
		t.Fatalf("dashboard redirect = %q, want /login", location)
	}
}

func TestProtectedRouteRejectsNonAdmin(t *testing.T) {
	srv := testServer(stubAuthenticator{err: backofficeauth.ErrForbidden})
	recorder := httptest.NewRecorder()
	srv.Router.ServeHTTP(recorder, authenticatedRequest(t, http.MethodGet, "/"))

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("dashboard status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestInvalidSessionRedirectsToLogin(t *testing.T) {
	srv := testServer(stubAuthenticator{err: errors.New("invalid session")})
	recorder := httptest.NewRecorder()
	srv.Router.ServeHTTP(recorder, authenticatedRequest(t, http.MethodGet, "/"))

	if recorder.Code != http.StatusSeeOther {
		t.Fatalf("dashboard status = %d, want %d", recorder.Code, http.StatusSeeOther)
	}
	if location := recorder.Header().Get("Location"); location != "/login" {
		t.Fatalf("dashboard redirect = %q, want /login", location)
	}
}

func TestDashboard(t *testing.T) {
	srv := testServer(stubAuthenticator{principal: authenticatedPrincipal()})
	recorder := httptest.NewRecorder()
	srv.Router.ServeHTTP(recorder, authenticatedRequest(t, http.MethodGet, "/"))
	if recorder.Code != http.StatusOK {
		t.Fatalf("dashboard status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	if !strings.Contains(body, "Leamout Backoffice") {
		t.Fatalf("dashboard response missing Backoffice title")
	}
	if !strings.Contains(body, `hx-get="/fragments/runtime-status"`) {
		t.Fatalf("dashboard response missing HTMX runtime action")
	}
	if !strings.Contains(body, `_="on click`) {
		t.Fatalf("dashboard response missing Hyperscript interaction")
	}
}

func TestStaticAssets(t *testing.T) {
	srv := testServer(nil)
	for _, path := range []string{"/static/css/tailwindcss.css", "/static/js/htmx.min.js", "/static/js/hyperscript.min.js"} {
		recorder := httptest.NewRecorder()
		srv.Router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", path, recorder.Code, http.StatusOK)
		}
		if recorder.Body.Len() < 1000 {
			t.Errorf("GET %s returned an unexpectedly small asset (%d bytes)", path, recorder.Body.Len())
		}
	}
}

func TestModulePages(t *testing.T) {
	tests := []struct{ path, want string }{
		{path: "/organizations", want: "Organizations"},
		{path: "/calls", want: "Calls"},
		{path: "/numbers", want: "Phone Numbers"},
		{path: "/trunks", want: "SIP Trunks"},
		{path: "/carrier-connections", want: "Carrier Connections"},
		{path: "/providers", want: "Managed Providers"},
		{path: "/commercial", want: "Commercial"},
	}

	srv := testServer(stubAuthenticator{principal: authenticatedPrincipal()})
	for _, tt := range tests {
		recorder := httptest.NewRecorder()
		srv.Router.ServeHTTP(recorder, authenticatedRequest(t, http.MethodGet, tt.path))
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", tt.path, recorder.Code, http.StatusOK)
		}
		if !strings.Contains(recorder.Body.String(), tt.want) {
			t.Errorf("GET %s response does not contain %q", tt.path, tt.want)
		}
	}
}
