package commercial

import (
	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/server/internal/platform/deployment"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSelfHostedExcludesCloudRoutes(t *testing.T) {
	module := New(nil, Dependencies{Deployment: deployment.Config{Mode: deployment.SelfHosted}})
	router := chi.NewRouter()
	auth := func(string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "forbidden", http.StatusForbidden)
			})
		}
	}
	RegisterRoutes(router, module, auth, func(next http.Handler) http.Handler { return next })
	for _, tc := range []struct {
		method, path string
		code         int
	}{
		{http.MethodPost, "/billing/checkout", 404},
		{http.MethodPost, "/billing/portal", 404},
		{http.MethodPost, "/webhooks/polar", 404},
		{http.MethodGet, "/plans", 404},
		{http.MethodGet, "/subscription", 404},
		{http.MethodGet, "/license", 404},
		{http.MethodGet, "/limits", 403},
		{http.MethodGet, "/runtime-usage", 403},
		{http.MethodGet, "/entitlements", 403},
	} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, nil))
		if rec.Code != tc.code {
			t.Errorf("%s %s: %d want %d", tc.method, tc.path, rec.Code, tc.code)
		}
	}
}
