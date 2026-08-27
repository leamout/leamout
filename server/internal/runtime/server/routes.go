package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/leamout/leamout/internal/identity/auth"
	"github.com/leamout/leamout/internal/identity/session"
	"github.com/leamout/leamout/internal/identity/users"
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
	})
}
