package commercial

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/server/internal/commercial/licensing"
)

func TestSelfHostedExcludesCloudRoutesAndProtectsLicenseStatus(t *testing.T) {
	public, _, _ := ed25519.GenerateKey(rand.Reader)
	license, err := licensing.NewService(licensing.Config{Mode: licensing.SelfHosted, DeploymentID: "deployment", File: filepath.Join(t.TempDir(), "missing"), PublicKeys: map[string]string{"issuer": base64.StdEncoding.EncodeToString(public)}})
	if err != nil {
		t.Fatal(err)
	}
	module := New(nil, Dependencies{License: license})
	router := chi.NewRouter()
	auth := func(string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Error(w, "forbidden", http.StatusForbidden) })
		}
	}
	RegisterRoutes(router, module, auth, func(h http.Handler) http.Handler { return h })
	for _, tc := range []struct {
		method, path string
		code         int
	}{
		{http.MethodPost, "/billing/checkout", 404}, {http.MethodPost, "/billing/portal", 404}, {http.MethodPost, "/webhooks/stripe", 404}, {http.MethodGet, "/plans", 404}, {http.MethodGet, "/subscription", 404},
		{http.MethodGet, "/license", 403}, {http.MethodGet, "/limits", 403}, {http.MethodGet, "/runtime-usage", 403}, {http.MethodGet, "/entitlements", 403},
	} {
		t.Run(tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
			if rec.Code != tc.code {
				t.Fatalf("status=%d want=%d", rec.Code, tc.code)
			}
		})
	}
}
