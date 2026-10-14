package commercial

import (
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/go-chi/chi/v5"
)

func TestCloudRoutesRequireOrganizationAuthorization(t *testing.T) {
    router := chi.NewRouter()
    module := New(nil)
    deny := func(string) func(http.Handler) http.Handler {
        return func(next http.Handler) http.Handler {
            return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                http.Error(w, "forbidden", http.StatusForbidden)
            })
        }
    }
    RegisterRoutes(router, module, deny, func(next http.Handler) http.Handler { return next })
    for _, path := range []string{"/billing/status", "/limits", "/runtime-usage", "/entitlements"} {
        rec := httptest.NewRecorder()
        router.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))
        if rec.Code != http.StatusForbidden {
            t.Errorf("%s: got %d, want 403", path, rec.Code)
        }
    }
}
