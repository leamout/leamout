package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBrowser(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, origin, contentType, fetchSite string
		status                                             int
	}{
		{"console login", "POST", "/v1/auth/logout", "https://console.example.com", "application/json", "same-origin", 204},
		{"legacy browser", "POST", "/v1/auth/logout", "https://console.example.com", "application/json; charset=utf-8", "", 204},
		{"foreign origin", "POST", "/v1/auth/logout", "https://evil.example.com", "application/json", "same-origin", 403},
		{"null origin", "POST", "/v1/auth/logout", "null", "application/json", "", 403},
		{"cross-site without origin", "DELETE", "/v1/sessions/current", "", "", "cross-site", 403},
		{"logout", "DELETE", "/v1/sessions/current", "https://console.example.com", "", "same-origin", 204},
		{"foreign logout", "DELETE", "/v1/sessions/current", "https://evil.example.com", "", "cross-site", 403},
		{"non-browser client", "POST", "/v1/auth/logout", "", "application/json", "", 204},
		{"form post", "POST", "/v1/auth/logout", "", "text/plain", "", 415},
		{"read", "GET", "/v1/users/me", "https://evil.example.com", "", "cross-site", 204},
		{"preflight", "OPTIONS", "/v1/auth/logout", "https://console.example.com", "", "cross-site", 204},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, nil)
			r.Header.Set("Origin", tc.origin)
			r.Header.Set("Content-Type", tc.contentType)
			r.Header.Set("Sec-Fetch-Site", tc.fetchSite)
			w := httptest.NewRecorder()
			Browser("https://console.example.com")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})).ServeHTTP(w, r)
			if w.Code != tc.status {
				t.Fatalf("status = %d, want %d", w.Code, tc.status)
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("API responses must not be cached")
			}
		})
	}
}
