package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/leamout/leamout/internal/identity/auth"
	"github.com/leamout/leamout/internal/identity/session"
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
	})
}
