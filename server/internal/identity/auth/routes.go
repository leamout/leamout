package auth

import "github.com/go-chi/chi/v5"

func RegisterRoutes(router chi.Router, handler *Handler) {
	router.Post("/identity/backchannel-logout", handler.BackchannelLogout)
	router.Route("/auth", func(r chi.Router) {
		r.Get("/login", handler.Login)
		r.Get("/callback", handler.Callback)
		r.Post("/logout", handler.Logout)
		r.Get("/account", handler.Account)
	})
}
