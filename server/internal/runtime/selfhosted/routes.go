package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/leamout/leamout/internal/identity"
	sharedmodules "github.com/leamout/leamout/internal/modules"
	"github.com/leamout/leamout/internal/telecom"
	"github.com/leamout/leamout/internal/tenancy"
)

func RegisterRoutes(r *chi.Mux, modules Modules) {
	organizationAccess := func(resource string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			requireAuthenticated := modules.OrganizationsContext.RequireAuthenticated(modules.Authn)
			requireAccess := modules.OrganizationsContext.RequireAccess(resource)
			return requireAuthenticated(modules.RateLimit.Handle(requireAccess(next)))
		}
	}
	sessionOrganizationAccess := func(resource string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			requireAccess := modules.OrganizationsContext.RequireAccess(resource)
			return modules.Authn.RequireSession(
				modules.OrganizationsContext.Require(modules.RateLimit.Handle(requireAccess(next))),
			)
		}
	}
	organizationContextAccess := func(resource string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			requireAccess := modules.OrganizationsContext.RequireAccess(resource)
			return modules.OrganizationsContext.Require(modules.RateLimit.Handle(requireAccess(next)))
		}
	}

	r.Route("/v1", func(r chi.Router) {
		identity.RegisterRoutes(r, modules.Identity, modules.Authn.RequireSession)
		tenancy.RegisterRoutes(
			r,
			modules.Tenancy,
			modules.Authn.RequireSession,
			organizationContextAccess,
			sessionOrganizationAccess,
		)
		telecom.RegisterRoutes(r, modules.Telecom, organizationAccess, modules.Shared.Idempotency.Middleware.Handle)
		sharedmodules.RegisterRoutes(r, modules.Shared, organizationAccess)
	})
}
