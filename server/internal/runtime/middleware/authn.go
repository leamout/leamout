package middleware

import (
	"net/http"

	"github.com/leamout/leamout/internal/security/authn"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/httputil"
)

const sessionCookieName = "leamout-session"

type AuthnMiddleware struct {
	resolver *authn.Resolver
}

func NewAuthnMiddleware(resolver *authn.Resolver) *AuthnMiddleware {
	return &AuthnMiddleware{
		resolver: resolver,
	}
}

func (m *AuthnMiddleware) RequireSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || cookie.Value == "" {
			httputil.Error(
				w,
				apperror.NewUnauthorized("authentication required"),
			)
			return
		}

		principal, err := m.resolver.Resolve(
			r.Context(),
			authn.CredentialInput{
				Type:  authn.CredentialSession,
				Value: cookie.Value,
			},
		)
		if err != nil {
			httputil.Error(
				w,
				apperror.NewUnauthorized("authentication required"),
			)
			return
		}

		ctx := authn.WithPrincipal(r.Context(), principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
