package organizations

import "github.com/go-chi/chi/v5"

func (h *Handler) Routes(router chi.Router) {
	router.Get("/organizations", h.index)
	router.Get("/organizations/{organization_id}", h.detail)
}
