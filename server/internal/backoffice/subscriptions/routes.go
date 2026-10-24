package subscriptions

import "github.com/go-chi/chi/v5"

func (h *Handler) Routes(router chi.Router) {
	router.Get("/subscriptions", h.index)
	router.Get("/subscriptions/{subscription_id}", h.detail)
}
