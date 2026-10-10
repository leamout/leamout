package commercial

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/server/internal/commercial/billing"
	"github.com/leamout/leamout/server/internal/commercial/entitlements"
	"github.com/leamout/leamout/server/internal/commercial/plans"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
)

func RegisterRoutes(
	router chi.Router,
	module *Module,
	organizationAccess func(string) func(http.Handler) http.Handler,
) {
	plans.RegisterRoutes(router, module.Plans.Handler)

	auth := organizationAccess("commercial")
	billing.RegisterRoutes(router, module.Billing.Handler, auth)
	subscriptions.RegisterRoutes(router, module.Subscriptions.Handler, auth)
	entitlements.RegisterRoutes(router, module.Entitlements.Handler, auth)
}

func RegisterPublicRoutes(router chi.Router, module *Module) {
	billing.RegisterStripeWebhookRoute(router, module.Billing.Handler)
}
