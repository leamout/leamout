package wallets

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, handler *Handler, auth, idempotency func(http.Handler) http.Handler) {
	router.With(auth, idempotency).Post("/wallets/{wallet_id}/topups", handler.CreateTopup)
}
