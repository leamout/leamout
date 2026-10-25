package numbers

import "github.com/go-chi/chi/v5"

func (h *Handler) Routes(router chi.Router) {
	router.Get("/numbers", h.index)
	router.Get("/numbers/{number_id}", h.detail)
}
