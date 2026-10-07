package platform

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/leamout/leamout/server/internal/platform/audit"
	"github.com/leamout/leamout/server/internal/platform/entitlements"
	"github.com/leamout/leamout/server/internal/platform/networking"
	"github.com/leamout/leamout/server/internal/platform/retention"
	"github.com/leamout/leamout/server/internal/platform/storage"
	"github.com/leamout/leamout/server/internal/platform/webhooks"
	"github.com/leamout/leamout/server/internal/security/scim"
)

func RegisterRoutes(
	router chi.Router,
	module *Module,
	organizationAccess func(string) func(http.Handler) http.Handler,
) {
	webhooks.RegisterRoutes(
		router,
		module.Webhooks.Handler,
		organizationAccess("webhooks"),
	)

	audit.RegisterRoutes(
		router,
		module.Audit.Handler,
		organizationAccess("audit"),
	)

	storage.RegisterRoutes(
		router,
		module.Storage.Handler,
		organizationAccess("storage"),
	)

	networking.RegisterRoutes(
		router,
		module.Networking.Handler,
		organizationAccess("networking"),
		module.Entitlements.Middleware.Require(entitlements.CapabilityPrivateNetworking),
	)

	scim.RegisterManagementRoutes(
		router,
		module.SCIM.Handler,
		organizationAccess("scim"),
		module.Entitlements.Middleware.Require(entitlements.CapabilitySCIM),
	)

	retention.RegisterRoutes(
		router,
		module.Retention.Handler,
		organizationAccess("retention"),
		module.Entitlements.Middleware.Require(entitlements.CapabilityRetentionPolicies),
	)
}
