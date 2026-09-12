package prepaid

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, handler *Handler, auth func(http.Handler) http.Handler) {
	router.With(auth).Get("/wallets", handler.List)
	router.With(auth).Get("/wallets/{wallet_id}", handler.Get)
	router.With(auth).Get("/wallets/{wallet_id}/ledger", handler.ListLedger)
}
