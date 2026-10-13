package limits

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, handler *Handler, auth func(http.Handler) http.Handler) {
	router.With(auth).Get("/limits", handler.Get)
	router.With(auth).Get("/runtime-usage", handler.RuntimeUsage)
}
