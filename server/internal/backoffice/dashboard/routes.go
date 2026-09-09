package dashboard

import "github.com/go-chi/chi/v5"

func (h *Handler) Routes(router chi.Router) {
	router.Get("/", h.index)
	router.Get("/fragments/runtime-status", h.runtimeStatus)
}
