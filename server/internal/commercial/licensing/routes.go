package licensing

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, handler *Handler, auth func(http.Handler) http.Handler) {
	router.Route("/licenses", func(r chi.Router) {
		r.Use(auth)
		r.Get("/", handler.List)
		r.Get("/{license_id}", handler.Get)
		r.Post("/", handler.Create)
		r.Get("/{license_id}/deployments", handler.ListDeployments)
		r.Post("/{license_id}/deployments", handler.ActivateDeployment)
		r.Post("/{license_id}/deployments/{deployment_id}/heartbeat", handler.HeartbeatDeployment)
		r.Delete("/{license_id}/deployments/{deployment_id}", handler.DeactivateDeployment)
	})
}
