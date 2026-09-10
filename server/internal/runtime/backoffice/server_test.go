package backoffice

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/security/authn"
)

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
		Assurance: authn.AssurancePassword,
	}
}

func authenticateAs(principal authn.Principal) accessMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := authn.WithPrincipal(r.Context(), principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func rejectAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func allowAdmin(next http.Handler) http.Handler {
	return next
}

func rejectAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Forbidden", http.StatusForbidden)
	})
}

func testServer(access ...accessMiddleware) *Server {
	return newServer(nil, newModules(nil), access...)
}

func TestHealth(t *testing.T) {
	srv := testServer(rejectAuthentication)
	recorder := httptest.NewRecorder()
	srv.Router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestProtectedRouteRequiresAuthentication(t *testing.T) {
	srv := testServer(rejectAuthentication)
	recorder := httptest.NewRecorder()
	srv.Router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("dashboard status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestProtectedRouteRejectsNonAdmin(t *testing.T) {
	srv := testServer(authenticateAs(authenticatedPrincipal()), rejectAdmin)
	recorder := httptest.NewRecorder()
	srv.Router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("dashboard status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestUnsafeRequestRejectsSiblingOrigin(t *testing.T) {
	handler := protectUnsafeRequests(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	recorder := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mutation", nil)
	r.Host = "backoffice.leamout.com"
	r.Header.Set("Origin", "https://api.leamout.com")
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("sibling-origin POST status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestUnsafeRequestAllowsSameOrigin(t *testing.T) {
	handler := protectUnsafeRequests(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	recorder := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mutation", nil)
	r.Host = "backoffice.leamout.com"
	r.Header.Set("Origin", "https://backoffice.leamout.com")
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("same-origin POST status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestDashboard(t *testing.T) {
	srv := testServer(authenticateAs(authenticatedPrincipal()), allowAdmin)
	recorder := httptest.NewRecorder()
	srv.Router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil))
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
	srv := testServer(rejectAuthentication)
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
		{path: "/users", want: "Users"},
		{path: "/organizations", want: "Organizations"},
		{path: "/calls", want: "Calls"},
		{path: "/numbers", want: "Phone Numbers"},
		{path: "/trunks", want: "SIP Trunks"},
		{path: "/carrier-connections", want: "Carrier Connections"},
		{path: "/providers", want: "Managed Providers"},
		{path: "/commercial", want: "Commercial"},
	}

	srv := testServer(authenticateAs(authenticatedPrincipal()), allowAdmin)
	for _, tt := range tests {
		recorder := httptest.NewRecorder()
		srv.Router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, tt.path, nil))
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", tt.path, recorder.Code, http.StatusOK)
		}
		if !strings.Contains(recorder.Body.String(), tt.want) {
			t.Errorf("GET %s response does not contain %q", tt.path, tt.want)
		}
	}
}
