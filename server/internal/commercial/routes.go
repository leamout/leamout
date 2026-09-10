package commercial

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/internal/commercial/catalog"
	"github.com/leamout/leamout/internal/commercial/licensing"
	commercialstate "github.com/leamout/leamout/internal/commercial/state"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

type Middleware func(http.Handler) http.Handler

type RouteMiddleware struct {
	Session            Middleware
	OrganizationAccess func(resource string) Middleware
	Idempotency        Middleware
}

// RegisterRoutes exposes Commercial HTTP surfaces while keeping authentication
// and tenant authorization policy owned by the runtime.
func RegisterRoutes(router chi.Router, module *Module, middleware RouteMiddleware) {
	catalog.RegisterRoutes(router, module.Catalog.Handler, middleware.Session)
	licensing.RegisterRoutes(
		router,
		module.Access.Licensing.Handler,
		middleware.OrganizationAccess("licensing"),
		middleware.Idempotency,
	)
	commercialstate.RegisterRoutes(
		router,
		module.State.Handler,
		middleware.OrganizationAccess("commercial-state"),
	)
	subscriptions.RegisterRoutes(
		router,
		module.Access.Subscriptions.Handler,
		middleware.OrganizationAccess("subscriptions"),
		middleware.Idempotency,
	)
	wallets.RegisterTopupRoutes(
		router,
		module.Money.TopupHandler,
		middleware.OrganizationAccess("billing"),
		middleware.Idempotency,
	)
}
