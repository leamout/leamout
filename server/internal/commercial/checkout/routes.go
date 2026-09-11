package checkout

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, handler *Handler, auth, idempotency func(http.Handler) http.Handler) {
	router.With(auth, idempotency).Post("/checkouts", handler.Create)
	router.With(auth).Get("/checkouts/{checkout_id}", handler.Get)
	router.With(auth, idempotency).Post("/checkouts/{checkout_id}/continue", handler.Continue)
}
