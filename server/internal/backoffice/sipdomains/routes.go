package sipdomains

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (h *Handler) Routes() http.Handler {
	router := chi.NewRouter()
	router.Get("/", h.index)
	router.Get("/{domain_id}", h.detail)
	return router
}
