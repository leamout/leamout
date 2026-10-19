package manifest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	router chi.Router,
	handler *Handler,
	auth func(http.Handler) http.Handler,
) {
	router.Route("/agent-packages", func(r chi.Router) {
		r.Use(auth)
		r.Post("/deploy", handler.Deploy)
	})
}
