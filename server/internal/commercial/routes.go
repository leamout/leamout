package commercial

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/internal/commercial/catalog"
	"github.com/leamout/leamout/internal/commercial/licensing"
	"github.com/leamout/leamout/internal/commercial/payments"
	commercialstate "github.com/leamout/leamout/internal/commercial/state"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/internal/commercial/topups"
)

// RegisterRoutes exposes Commercial HTTP routes. Authentication, organization
// authorization, and idempotency remain runtime middleware concerns.
func RegisterRoutes(
	router chi.Router,
	module *Module,
	requireSession func(http.Handler) http.Handler,
	organizationAccess func(string) func(http.Handler) http.Handler,
	idempotency func(http.Handler) http.Handler,
) {
	catalog.RegisterRoutes(router, module.Catalog.Handler, requireSession)
	licensing.RegisterRoutes(
		router,
		module.Access.Licenses.Handler,
		organizationAccess("licensing"),
		idempotency,
	)
	commercialstate.RegisterRoutes(
		router,
		module.Access.State.Handler,
		organizationAccess("commercial-state"),
	)
	subscriptions.RegisterRoutes(
		router,
		module.Access.Subscriptions.Handler,
		organizationAccess("subscriptions"),
		idempotency,
	)
	topups.RegisterRoutes(
		router,
		module.Purchase.Topups.Handler,
		organizationAccess("billing"),
		idempotency,
	)
	payments.RegisterRoutes(router, module.Payments.Handler)
}
