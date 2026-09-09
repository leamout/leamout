package auth

import "github.com/go-chi/chi/v5"

func (h *Handler) RegisterPublicRoutes(router chi.Router) {
	router.Get("/login", h.loginPage)
	router.Post("/login", h.login)
}

func (h *Handler) RegisterProtectedRoutes(router chi.Router) {
	router.Post("/logout", h.logout)
}
