package billing

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	router chi.Router,
	handler *Handler,
	auth func(http.Handler) http.Handler,
) {
	router.Route("/billing", func(r chi.Router) {
		r.Use(auth)
		r.Post("/checkout", handler.Checkout)
		r.Post("/portal", handler.Portal)
	})
}

func RegisterStripeWebhookRoute(router chi.Router, handler *Handler) {
	router.Post("/webhooks/stripe", handler.StripeWebhook)
}
