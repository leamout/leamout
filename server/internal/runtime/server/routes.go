package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/leamout/leamout/internal/identity/auth"
	"github.com/leamout/leamout/internal/identity/session"
	"github.com/leamout/leamout/internal/identity/users"
	"github.com/leamout/leamout/internal/modules/webhooks"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/conferences"
	"github.com/leamout/leamout/internal/telecom/recordings"
	"github.com/leamout/leamout/internal/telecom/sip_domains"
	"github.com/leamout/leamout/internal/telecom/voice"
	"github.com/leamout/leamout/internal/tenancy/credentials"
	"github.com/leamout/leamout/internal/tenancy/members"
	"github.com/leamout/leamout/internal/tenancy/organization"
)

func RegisterRoutes(r *chi.Mux, modules Modules) {
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	r.Route("/v1", func(r chi.Router) {
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
		)
		members.RegisterRoutes(
			r,
			modules.Members.Handler,
			modules.Authn.RequireSession,
		)
		credentials.RegisterRoutes(
			r,
			modules.Credentials.Handler,
			modules.Authn.RequireSession,
		)
		voice.RegisterRoutes(
			r,
			modules.Voice.Handler,
			modules.OrganizationsContext.RequireAuthenticated(modules.Authn),
		)
		calls.RegisterRoutes(
			r,
			modules.Calls.Handler,
			modules.OrganizationsContext.RequireAuthenticated(modules.Authn),
		)
		recordings.RegisterRoutes(
			r,
			modules.Recordings.Handler,
			modules.OrganizationsContext.RequireAuthenticated(modules.Authn),
		)
		sip_domains.RegisterRoutes(
			r,
			modules.SIPDomains.Handler,
			modules.OrganizationsContext.RequireAuthenticated(modules.Authn),
		)
		webhooks.RegisterRoutes(
			r,
			modules.Webhooks.Handler,
			modules.OrganizationsContext.RequireAuthenticated(modules.Authn),
		)
		conferences.RegisterRoutes(
			r,
			modules.Conferences.Handler,
			modules.OrganizationsContext.RequireAuthenticated(modules.Authn),
		)
	})
}
