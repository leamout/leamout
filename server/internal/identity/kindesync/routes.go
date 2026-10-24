package kindesync

import "github.com/go-chi/chi/v5"

func RegisterRoutes(router chi.Router, handler *Handler) {
	router.Post("/identity/kinde/webhook", handler.Receive)
}
