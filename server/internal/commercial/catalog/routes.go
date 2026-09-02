package catalog

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, handler *Handler, auth func(http.Handler) http.Handler) {
	router.Route("/products", func(r chi.Router) {
		r.Use(auth)
		r.Get("/", handler.ListProducts)
		r.Get("/{product_id}", handler.GetProduct)
		r.Get("/{product_id}/plans", handler.ListPlans)
	})

	router.Route("/plans", func(r chi.Router) {
		r.Use(auth)
		r.Get("/{plan_id}", handler.GetPlan)
		r.Get("/{plan_id}/prices", handler.ListPrices)
	})
}
