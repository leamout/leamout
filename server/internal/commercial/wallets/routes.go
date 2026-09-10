package wallets

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	router chi.Router,
	handler *TopupHandler,
	auth func(http.Handler) http.Handler,
	idempotency func(http.Handler) http.Handler,
) {
	router.With(auth, idempotency).Post("/wallets/{wallet_id}/topups", handler.Create)
	router.With(auth).Get("/checkouts/{checkout_id}", handler.Get)
	router.With(auth, idempotency).Post("/checkouts/{checkout_id}/continue", handler.Continue)
	router.Post("/payment-webhooks/{provider}", handler.Webhook)
}
