package billing

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	router chi.Router,
	handler *Handler,
	auth func(http.Handler) http.Handler,
	idempotency func(http.Handler) http.Handler,
) {
	router.Route("/billing", func(r chi.Router) {
		r.Use(auth)
		r.With(idempotency).Post("/checkout", handler.Checkout)
		r.Post("/portal", handler.Portal)
	})

	// Polar authenticates the raw body with Standard Webhooks headers.
	router.Post("/webhooks/polar", handler.PolarWebhook)
}
