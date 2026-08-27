package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/security/authn"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/httputil"
)

const organizationIDHeader = "X-Organization-ID"

type organizationContext struct {
	ID uuid.UUID
}

type OrganizationMiddleware struct{}

func NewOrganizationMiddleware() *OrganizationMiddleware {
	return &OrganizationMiddleware{}
}

func (m *OrganizationMiddleware) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := authn.PrincipalFromContext(r.Context())
		if !ok {
			httputil.Error(w, apperror.NewUnauthorized("authentication required"))
			return
		}

		organizationID, ok := resolveOrganizationID(r, principal)
		if !ok {
			httputil.Error(w, apperror.NewBadRequest("organization context required"))
			return
		}

		ctx := withOrganizationContext(r.Context(), organizationContext{ID: organizationID})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func resolveOrganizationID(r *http.Request, principal authn.Principal) (uuid.UUID, bool) {
	if principal.Credential.Type == authn.CredentialOrganizationToken {
		if principal.OrganizationID == uuid.Nil {
			return uuid.Nil, false
		}

		return principal.OrganizationID, true
	}

	value := r.Header.Get(organizationIDHeader)
	if value == "" {
		return uuid.Nil, false
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, false
	}

	return id, true
}

type organizationContextKey struct{}

func withOrganizationContext(ctx context.Context, organization organizationContext) context.Context {
	return context.WithValue(ctx, organizationContextKey{}, organization)
}

func OrganizationIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	organization, ok := ctx.Value(organizationContextKey{}).(organizationContext)
	if !ok || organization.ID == uuid.Nil {
		return uuid.Nil, false
	}

	return organization.ID, true
}
