package agents

import "github.com/go-chi/chi/v5"

func (h *Handler) Routes(router chi.Router) {
	router.Get("/agents", h.index)
	router.Get("/agents/{agent_id}", h.detail)
}
