package backoffice

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/observability/logging"
	"github.com/leamout/leamout/server/internal/security/authn"
)

func requirePlatformAdmin(queries *sqlc.Queries, logger *logging.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := authn.PrincipalFromContext(r.Context())
			if !ok || principal.Credential.Type != authn.CredentialSession {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
			userID, ok := principal.UserID()
			if !ok {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
			user, err := queries.GetUserByID(r.Context(), userID)
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}
			if err != nil {
				logger.Error(r.Context(), "resolve backoffice operator", "error", err)
				http.Error(w, "Unable to verify operator access", http.StatusServiceUnavailable)
				return
			}
			if !user.IsPlatformAdmin {
				http.Error(w, "Platform administrator access required", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; object-src 'none'; base-uri 'self'; frame-ancestors 'none'; form-action 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
