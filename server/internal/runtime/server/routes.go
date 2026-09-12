package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/leamout/leamout/internal/commercial"
	"github.com/leamout/leamout/internal/identity"
	"github.com/leamout/leamout/internal/modules/audit"
	"github.com/leamout/leamout/internal/modules/webhooks"
	providerdiagnostics "github.com/leamout/leamout/internal/platform/provider_diagnostics"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/carriers"
	"github.com/leamout/leamout/internal/telecom/conferences"
	"github.com/leamout/leamout/internal/telecom/numbers"
	"github.com/leamout/leamout/internal/telecom/realtime"
	"github.com/leamout/leamout/internal/telecom/recordings"
	"github.com/leamout/leamout/internal/telecom/sip_domains"
	"github.com/leamout/leamout/internal/telecom/subscribers"
	"github.com/leamout/leamout/internal/telecom/trunks"
	"github.com/leamout/leamout/internal/telecom/voice"
	"github.com/leamout/leamout/internal/tenancy"
)

func RegisterRoutes(r *chi.Mux, modules Modules) {
	if modules.Edge.Handler != nil {
		r.Post("/internal/v1/sip-edge/authorize", modules.Edge.Handler.Admit)
	}
	if modules.Wholesale.Handler != nil {
		r.Post("/internal/v1/provider-cdrs/reconcile", modules.Wholesale.Handler.Reconcile)
	}
	providerdiagnostics.RegisterRoutes(r, modules.ProviderDiagnostics.Handler)

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
		commercial.RegisterRoutes(
			r,
			modules.Commercial,
			modules.Authn.RequireSession,
			organizationAccess,
			modules.Idempotency.Middleware.Handle,
		)
		identity.RegisterRoutes(r, modules.Identity, modules.Authn.RequireSession)
		tenancy.RegisterRoutes(
			r,
			modules.Tenancy,
			modules.Authn.RequireSession,
			organizationContextAccess,
			sessionOrganizationAccess,
		)
		voice.RegisterRoutes(r, modules.Voice.Handler, organizationAccess("voice-applications"))
		calls.RegisterRoutes(r, modules.Calls.Handler, organizationAccess("calls"))
		recordings.RegisterRoutes(r, modules.Recordings.Handler, organizationAccess("recordings"))
		subscribers.RegisterRoutes(r, modules.Subscribers.Handler, organizationAccess("subscribers"))
		numbers.RegisterRoutes(
			r,
			modules.Numbers.Handler,
			organizationAccess("numbers"),
			modules.Idempotency.Middleware.Handle,
		)
		sip_domains.RegisterRoutes(r, modules.SIPDomains.Handler, organizationAccess("sip-domains"))
		trunks.RegisterRoutes(r, modules.Trunks.Handler, organizationAccess("trunks"))
		carriers.RegisterRoutes(r, modules.Carriers.Handler, organizationAccess("carriers"))
		webhooks.RegisterRoutes(r, modules.Webhooks.Handler, organizationAccess("webhooks"))
		audit.RegisterRoutes(r, modules.Audit.Handler, organizationAccess("audit"))
		conferences.RegisterRoutes(r, modules.Conferences.Handler, organizationAccess("conferences"))
		realtime.RegisterRoutes(r, modules.Realtime.Handler, organizationAccess("realtime"))
	})
}
