package number_orders

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, handler *Handler, auth func(http.Handler) http.Handler) {
	router.Route("/number-orders", func(r chi.Router) {
		r.Use(auth)
		r.Get("/", handler.List)
		r.Get("/{number_order_id}", handler.Get)
	})
}
