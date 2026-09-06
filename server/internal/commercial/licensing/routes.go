package licensing

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(
	router chi.Router,
	handler *Handler,
	auth func(http.Handler) http.Handler,
	idempotency func(http.Handler) http.Handler,
) {
	router.Route("/licenses", func(r chi.Router) {
		r.Use(auth)
		r.Get("/", handler.List)
		r.Get("/{license_id}", handler.Get)
		r.Get("/{license_id}/deployments", handler.ListDeployments)
		r.With(idempotency).Post("/{license_id}/deployments", handler.ActivateDeployment)
		r.With(idempotency).Post("/{license_id}/deployments/{deployment_id}/managed-carrier/enrollment", handler.EnrollManagedCarrier)
		r.With(idempotency).Post("/{license_id}/deployments/{deployment_id}/heartbeat", handler.HeartbeatDeployment)
		r.With(idempotency).Delete("/{license_id}/deployments/{deployment_id}", handler.DeactivateDeployment)
	})
}
