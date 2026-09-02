package subscriptions

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, handler *Handler, auth func(http.Handler) http.Handler) {
	router.Route("/subscriptions", func(r chi.Router) {
		r.Use(auth)
		r.Get("/", handler.List)
		r.Get("/{subscription_id}", handler.Get)
		r.Patch("/{subscription_id}", handler.Update)
		r.Post("/{subscription_id}/cancel", handler.Cancel)
	})
}
