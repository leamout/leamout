package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/leamout/leamout/internal/commercial/catalog"
	"github.com/leamout/leamout/internal/commercial/licensing"
	commercialstate "github.com/leamout/leamout/internal/commercial/state"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/internal/identity/auth"
	"github.com/leamout/leamout/internal/identity/session"
	"github.com/leamout/leamout/internal/identity/users"
	"github.com/leamout/leamout/internal/modules/audit"
	"github.com/leamout/leamout/internal/modules/webhooks"
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
	"github.com/leamout/leamout/internal/tenancy/credentials"
	"github.com/leamout/leamout/internal/tenancy/members"
	"github.com/leamout/leamout/internal/tenancy/organization"
)

func RegisterRoutes(r *chi.Mux, modules Modules) {
	organizationAccess := func(resource string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return modules.OrganizationsContext.RequireAuthenticated(modules.Authn)(
				modules.OrganizationsContext.RequireAccess(resource)(next),
			)
		}
	}
	sessionOrganizationAccess := func(resource string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return modules.Authn.RequireSession(
				modules.OrganizationsContext.Require(
					modules.OrganizationsContext.RequireAccess(resource)(next),
				),
			)
		}
	}
	organizationContextAccess := func(resource string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return modules.OrganizationsContext.Require(
				modules.OrganizationsContext.RequireAccess(resource)(next),
			)
		}
	}

	r.Route("/v1", func(r chi.Router) {
		catalog.RegisterRoutes(
			r,
			modules.Catalog.Handler,
			modules.Authn.RequireSession,
		)
		licensing.RegisterRoutes(
			r,
			modules.Licensing.Handler,
			organizationAccess("licensing"),
			modules.Idempotency.Middleware.Handle,
		)
		commercialstate.RegisterRoutes(
			r,
			modules.CommercialState.Handler,
			organizationAccess("commercial-state"),
		)
		subscriptions.RegisterRoutes(
			r,
			modules.Subscriptions.Handler,
			organizationAccess("subscriptions"),
			modules.Idempotency.Middleware.Handle,
		)
		auth.RegisterRoutes(
			r,
			modules.Auth.Handler,
			modules.Authn.RequireSession,
		)
		session.RegisterRoutes(
			r,
			modules.Session.Handler,
			modules.Authn.RequireSession,
		)
		users.RegisterRoutes(
			r,
			modules.Users.Handler,
			modules.Authn.RequireSession,
		)
		organization.RegisterRoutes(
			r,
			modules.Organizations.Handler,
			modules.Authn.RequireSession,
			organizationContextAccess("organization"),
		)
		members.RegisterRoutes(
			r,
			modules.Members.Handler,
			sessionOrganizationAccess("members"),
		)
		credentials.RegisterRoutes(
			r,
			modules.Credentials.Handler,
			// Credential lifecycle operations stay session-only. In particular,
			// organization tokens must not be able to mint or elevate tokens.
			sessionOrganizationAccess("credentials"),
		)
		voice.RegisterRoutes(
			r,
			modules.Voice.Handler,
			organizationAccess("voice-applications"),
		)
		calls.RegisterRoutes(
			r,
			modules.Calls.Handler,
			organizationAccess("calls"),
		)
		recordings.RegisterRoutes(
			r,
			modules.Recordings.Handler,
			organizationAccess("recordings"),
		)
		subscribers.RegisterRoutes(
			r,
			modules.Subscribers.Handler,
			organizationAccess("subscribers"),
		)
		numbers.RegisterRoutes(
			r,
			modules.Numbers.Handler,
			organizationAccess("numbers"),
		)
		sip_domains.RegisterRoutes(
			r,
			modules.SIPDomains.Handler,
			organizationAccess("sip-domains"),
		)
		trunks.RegisterRoutes(
			r,
			modules.Trunks.Handler,
			organizationAccess("trunks"),
		)
		carriers.RegisterRoutes(
			r,
			modules.Carriers.Handler,
			organizationAccess("carriers"),
		)
		webhooks.RegisterRoutes(
			r,
			modules.Webhooks.Handler,
			organizationAccess("webhooks"),
		)
		audit.RegisterRoutes(
			r,
			modules.Audit.Handler,
			organizationAccess("audit"),
		)
		conferences.RegisterRoutes(
			r,
			modules.Conferences.Handler,
			organizationAccess("conferences"),
		)
		realtime.RegisterRoutes(
			r,
			modules.Realtime.Handler,
			organizationAccess("realtime"),
		)
	})
}
