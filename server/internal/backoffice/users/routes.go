package users

import "github.com/go-chi/chi/v5"

func (h *Handler) Routes(router chi.Router) {
	router.Get("/users", h.index)
	router.Get("/users/{user_id}", h.detail)
}
