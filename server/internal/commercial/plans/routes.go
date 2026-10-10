package plans

import "github.com/go-chi/chi/v5"

func RegisterRoutes(router chi.Router, handler *Handler) {
	router.Get("/plans", handler.List)
}
