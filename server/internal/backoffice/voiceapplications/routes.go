package voiceapplications

import (
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (h *Handler) Routes() http.Handler {
	router := chi.NewRouter()
	router.Get("/", h.index)
	router.Get("/{application_id}", h.detail)
	return router
}
