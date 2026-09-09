package backoffice

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealth(t *testing.T) {
	srv := New()
	recorder := httptest.NewRecorder()
	srv.Router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestDashboard(t *testing.T) {
	srv := New()
	recorder := httptest.NewRecorder()
	srv.Router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))
	response := recorder.Result()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read dashboard response: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("dashboard status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if !strings.Contains(string(body), "Leamout Backoffice") {
		t.Fatalf("dashboard response missing Backoffice title")
	}
	if !strings.Contains(string(body), `hx-get="/fragments/runtime-status"`) {
		t.Fatalf("dashboard response missing HTMX runtime action")
	}
	if !strings.Contains(string(body), `_="on click`) {
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
		srv.Router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", path, recorder.Code, http.StatusOK)
		}
		if recorder.Body.Len() < 1000 {
			t.Errorf("GET %s returned an unexpectedly small asset (%d bytes)", path, recorder.Body.Len())
		}
	}
}
