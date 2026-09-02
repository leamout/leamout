package middleware

import (
	"crypto/subtle"
	"net/http"
	"strings"
)

// RequireOperatorKey protects machine-to-machine operator endpoints with a
// dedicated bearer credential. An empty configured key fails closed.
func RequireOperatorKey(configuredKey string) func(http.Handler) http.Handler {
	configuredKey = strings.TrimSpace(configuredKey)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			provided, bearer := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			provided = strings.TrimSpace(provided)
			if configuredKey == "" || !bearer || provided == "" ||
				subtle.ConstantTimeCompare([]byte(provided), []byte(configuredKey)) != 1 {
				w.Header().Set("WWW-Authenticate", `Bearer realm="operator"`)
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
