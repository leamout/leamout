package invitations

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, handler *Handler, requireSession, requireMembers func(http.Handler) http.Handler) {
	router.Route("/organizations/{organization_id}/invitations", func(r chi.Router) {
		r.Use(requireMembers)
		r.Post("/", handler.Create)
		r.Get("/", handler.List)
		r.Delete("/{invitation_id}", handler.Revoke)
	})
	router.With(requireSession).Post("/invitations/accept", handler.Accept)
}
