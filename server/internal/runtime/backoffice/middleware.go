package backoffice

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/security/authn"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/httputil"
)

type userLookup interface {
	GetUserByID(context.Context, uuid.UUID) (sqlc.User, error)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; object-src 'none'; base-uri 'self'; frame-ancestors 'none'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		next.ServeHTTP(w, r)
	})
}

func requirePlatformAdmin(users userLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := authn.UserIDFromContext(r.Context())
			if !ok || users == nil {
				httputil.Error(w, apperror.NewUnauthorized("authentication required"))
				return
			}

			user, err := users.GetUserByID(r.Context(), userID)
			if err != nil {
				httputil.Error(w, apperror.NewUnauthorized("authentication required"))
				return
			}
			if !user.IsPlatformAdmin {
				httputil.Error(w, apperror.NewForbidden("backoffice access forbidden"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func protectUnsafeRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isSafeMethod(r.Method) || hasSameOrigin(r) {
			next.ServeHTTP(w, r)
			return
		}

		httputil.Error(w, apperror.NewForbidden("cross-site request forbidden"))
	})
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	default:
		return false
	}
}

func hasSameOrigin(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))
	if origin == "" {
		return false
	}

	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return false
	}

	return strings.EqualFold(parsed.Host, r.Host)
}
