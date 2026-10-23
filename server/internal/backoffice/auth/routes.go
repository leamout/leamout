package auth

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) Routes(router chi.Router, originProtection func(http.Handler) http.Handler) {
	router.Get("/log-in", h.loginPage)
	router.Get("/auth/login", h.identity.Login)
	router.Get("/auth/callback", h.identity.Callback)
	router.With(originProtection).Post("/auth/logout", h.identity.BackofficeLogout)
}
