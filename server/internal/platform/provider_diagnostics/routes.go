package provider_diagnostics

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, handler *Handler) {
	if handler == nil {
		return
	}
	r.Get("/internal/v1/operator/provider-diagnostics", handler.Snapshot)
}
