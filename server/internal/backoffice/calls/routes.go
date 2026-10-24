package calls

import "github.com/go-chi/chi/v5"

func (h *Handler) Routes(router chi.Router) {
	router.Get("/calls", h.index)
	router.Get("/calls/{call_id}", h.detail)
}
