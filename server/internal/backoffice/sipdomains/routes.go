package sipdomains

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) Routes() http.Handler {
	router := chi.NewRouter()
	router.Get("/", h.index)
	router.Get("/{domain_id}", h.detail)
	return router
}
