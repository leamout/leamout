package commercial

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/internal/commercial/catalog"
	checkout "github.com/leamout/leamout/internal/commercial/checkout"
	"github.com/leamout/leamout/internal/commercial/licensing"
	"github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

func RegisterRoutes(
	router chi.Router,
	module *Module,
	requireSession func(http.Handler) http.Handler,
	organizationAccess func(string) func(http.Handler) http.Handler,
	idempotency func(http.Handler) http.Handler,
) {
	catalog.RegisterRoutes(router, module.Catalog.Handler, requireSession)
	licensing.RegisterRoutes(router, module.Licensing.Handler, organizationAccess("licensing"), idempotency)
	wallets.RegisterRoutes(router, module.Wallets.Handler, organizationAccess("billing"))
	checkout.RegisterRoutes(router, module.Billing.Checkouts.Handler, organizationAccess("billing"), idempotency)
	payments.RegisterRoutes(router, module.Billing.Payments.Handler)
}
