package payments

import "github.com/go-chi/chi/v5"

func RegisterRoutes(router chi.Router, handler *Handler) {
	router.Post("/payment-webhooks/{provider}", handler.Webhook)
}
