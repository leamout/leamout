package organization

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	router chi.Router,
	handler *Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	router.Route("/organizations", func(r chi.Router) {
		r.Use(authMiddleware)

		r.Post("/", handler.Create)
		r.Get("/", handler.List)
		r.Get("/{organization_id}", handler.Get)
		r.Patch("/{organization_id}", handler.Update)
		r.Delete("/{organization_id}", handler.Delete)
	})
}
