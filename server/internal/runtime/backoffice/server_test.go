package backoffice

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	srv := New()
	recorder := httptest.NewRecorder()
	srv.Router.ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestDashboard(t *testing.T) {
	srv := New()
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
	srv := New()
	for _, path := range []string{
		"/static/css/custom.css",
		"/static/js/htmx.min.js",
		"/static/js/hyperscript.min.js",
	} {
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
	tests := []struct {
		path string
		want string
	}{
		{path: "/organizations", want: "Acme Communications"},
		{path: "/calls", want: "call_01JQ8YN7"},
	}

	srv := New()
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
