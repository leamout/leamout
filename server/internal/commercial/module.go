package commercial

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/commercial/catalog"
	"github.com/leamout/leamout/internal/commercial/checkout"
	"github.com/leamout/leamout/internal/commercial/entitlements"
	"github.com/leamout/leamout/internal/commercial/licensing"
	"github.com/leamout/leamout/internal/commercial/orders"
	"github.com/leamout/leamout/internal/commercial/payments"
	commercialstate "github.com/leamout/leamout/internal/commercial/state"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/internal/commercial/usage"
	"github.com/leamout/leamout/internal/commercial/wallets"
	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
)

// Module is the composition boundary for Leamout's Commercial domain.
// Runtime and telecom code should depend on this module rather than assembling
// Commercial subdomains independently.
type Module struct {
	Catalog  CatalogModule
	Purchase PurchaseModule
	Access   AccessModule
	Usage    UsageModule
	Prepaid  PrepaidModule
	Payments PaymentsModule

	// Deprecated: use Prepaid. Kept temporarily for runtime migration.
	Money PrepaidModule
	// Deprecated: use Access.State. Kept temporarily for runtime migration.
	State StateModule
}

type CatalogModule struct {
	Repository *catalog.Repository
	Service    *catalog.Service
	Handler    *catalog.Handler
}

type PurchaseModule struct {
	Checkouts *checkout.Repository
	Orders    *orders.Repository
}

type AccessModule struct {
	Subscriptions SubscriptionsModule
	Licenses      LicensesModule
	Entitlements  EntitlementsModule
	State         StateModule
}

type SubscriptionsModule struct {
	Repository *subscriptions.Repository
	Service    *subscriptions.Service
	Handler    *subscriptions.Handler
}

type LicensesModule struct {
	Repository *licensing.Repository
	Service    *licensing.Service
	Handler    *licensing.Handler
}

type EntitlementsModule struct {
	Repository *entitlements.Repository
	Service    *entitlements.Service
}

type StateModule struct {
	Service *commercialstate.Service
	Handler *commercialstate.Handler
}

type UsageModule struct {
	Repository *usage.Repository
	Service    *usage.Service
}

type PrepaidModule struct {
	Wallets      *wallets.Repository
	TopupService *wallets.TopupService
	TopupHandler *wallets.TopupHandler
}

type PaymentsModule struct {
	Repository *payments.Repository
}

// New composes the Commercial domain from its durable submodules. Payment
// providers are registered after construction so provider adapters remain
// outside Commercial state and are configured by the runtime.
func New(db *pgxpool.Pool) *Module {
	catalogRepository := catalog.NewRepository(db)
	catalogService := catalog.NewService(catalogRepository)

	subscriptionsRepository := subscriptions.NewRepository(db)
	subscriptionsService := subscriptions.NewService(
		subscriptionsRepository,
		catalogService,
	)

	entitlementsRepository := entitlements.NewRepository(db)
	entitlementsService := entitlements.NewService(
		entitlementsRepository,
		subscriptionsService,
	)

	commercialStateService := commercialstate.NewService(
		subscriptionsService,
		entitlementsService,
	)

	licensingRepository := licensing.NewRepository(db)
	licensingService := licensing.NewService(
		licensingRepository,
		commercialStateService,
	)

	usageRepository := usage.NewRepository(db)
	usageService := usage.NewService(usageRepository)

	walletRepository := wallets.NewRepository(db)
	checkoutRepository := checkout.NewRepository(db)
	orderRepository := orders.NewRepository(db)
	paymentRepository := payments.NewRepository(db)
	topupService := wallets.NewTopupService(
		walletRepository,
		checkoutRepository,
		paymentRepository,
		walletRepository,
		map[string]paymentprovider.Provider{},
	)
	topupHandler := wallets.NewTopupHandler(topupService)
	stateModule := StateModule{
		Service: commercialStateService,
		Handler: commercialstate.NewHandler(commercialStateService),
	}
	prepaidModule := PrepaidModule{
		Wallets:      walletRepository,
		TopupService: topupService,
		TopupHandler: topupHandler,
	}

	return &Module{
		Catalog: CatalogModule{
			Repository: catalogRepository,
			Service:    catalogService,
			Handler:    catalog.NewHandler(catalogService),
		},
		Purchase: PurchaseModule{
			Checkouts: checkoutRepository,
			Orders:    orderRepository,
		},
		Access: AccessModule{
			Subscriptions: SubscriptionsModule{
				Repository: subscriptionsRepository,
				Service:    subscriptionsService,
				Handler:    subscriptions.NewHandler(subscriptionsService),
			},
			Licenses: LicensesModule{
				Repository: licensingRepository,
				Service:    licensingService,
				Handler:    licensing.NewHandler(licensingService),
			},
			Entitlements: EntitlementsModule{
				Repository: entitlementsRepository,
				Service:    entitlementsService,
			},
			State: stateModule,
		},
		Usage: UsageModule{
			Repository: usageRepository,
			Service:    usageService,
		},
		Prepaid: prepaidModule,
		Payments: PaymentsModule{
			Repository: paymentRepository,
		},
		Money: prepaidModule,
		State: stateModule,
	}
}
