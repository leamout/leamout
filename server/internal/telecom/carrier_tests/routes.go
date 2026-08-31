package carrier_tests

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, handler *Handler, auth func(http.Handler) http.Handler) {
	router.With(auth).Post("/carrier-connections/{carrier_connection_id}/test-calls", handler.Create)
	router.With(auth).Get("/carrier-connections/{carrier_connection_id}/test-calls", handler.List)
}
