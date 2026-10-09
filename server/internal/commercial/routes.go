package commercial

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/server/internal/commercial/entitlements"
	"github.com/leamout/leamout/server/internal/commercial/plans"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
)

func RegisterRoutes(
	router chi.Router,
	module *Module,
	organizationAccess func(string) func(http.Handler) http.Handler,
) {
	router.Route("/commercial", func(r chi.Router) {
		plans.RegisterRoutes(r, module.Plans.Handler)

		auth := organizationAccess("commercial")
		subscriptions.RegisterRoutes(r, module.Subscriptions.Handler, auth)
		entitlements.RegisterRoutes(r, module.Entitlements.Handler, auth)
	})
}
