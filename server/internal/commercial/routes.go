package commercial

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/server/internal/commercial/billing"
	"github.com/leamout/leamout/server/internal/commercial/entitlements"
	"github.com/leamout/leamout/server/internal/commercial/licensing"
	"github.com/leamout/leamout/server/internal/commercial/limits"
	"github.com/leamout/leamout/server/internal/commercial/plans"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
)

func RegisterRoutes(
	router chi.Router,
	module *Module,
	organizationAccess func(string) func(http.Handler) http.Handler,
	idempotency func(http.Handler) http.Handler,
) {
	limits.RegisterRoutes(router, limits.NewHandler(module.Limits.Service), organizationAccess("commercial"))
	if module.License.SelfHosted() {
		licensing.RegisterRoutes(router, licensing.NewHandler(module.License), organizationAccess("commercial"))
		entitlements.RegisterRoutes(router, module.Entitlements.Handler, organizationAccess("commercial"))
		return
	}
	plans.RegisterRoutes(router, module.Plans.Handler)

	auth := organizationAccess("commercial")
	billing.RegisterRoutes(router, module.Billing.Handler, auth, idempotency)
	subscriptions.RegisterRoutes(router, module.Subscriptions.Handler, auth)
	entitlements.RegisterRoutes(router, module.Entitlements.Handler, auth)
}
