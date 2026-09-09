package backoffice

import (
	"context"
	"errors"
	"net/http"

	backofficeauth "github.com/leamout/leamout/internal/backoffice/auth"
	"github.com/leamout/leamout/internal/security/authn"
)

type authenticator interface {
	Authenticate(context.Context, string) (authn.Principal, error)
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

func requireBackoffice(authentication authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(backofficeauth.CookieName)
			if err != nil || cookie.Value == "" || authentication == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			principal, err := authentication.Authenticate(r.Context(), cookie.Value)
			if err != nil {
				if errors.Is(err, backofficeauth.ErrForbidden) {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				backofficeauth.ClearCookie(w)
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := authn.WithPrincipal(r.Context(), principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
