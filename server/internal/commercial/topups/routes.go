package topups

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, handler *Handler, auth, idempotency func(http.Handler) http.Handler) {
	router.With(auth, idempotency).Post("/wallets/{wallet_id}/topups", handler.Create)
	router.With(auth).Get("/checkout-orders/{checkout_order_id}", handler.Get)
	router.With(auth, idempotency).Post("/checkout-orders/{checkout_order_id}/continue", handler.Continue)
	router.Post("/payment-webhooks/{provider}", handler.Webhook)
}
