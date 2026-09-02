package state

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, handler *Handler, auth func(http.Handler) http.Handler) {
	router.With(auth).Get("/organizations/{organization_id}/commercial-state", handler.GetOrganizationState)
}
