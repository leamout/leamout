package commercial

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/commercial/catalog"
	"github.com/leamout/leamout/internal/commercial/checkout"
	"github.com/leamout/leamout/internal/commercial/entitlements"
	"github.com/leamout/leamout/internal/commercial/licensing"
	"github.com/leamout/leamout/internal/commercial/metering"
	"github.com/leamout/leamout/internal/commercial/payments"
	commercialstate "github.com/leamout/leamout/internal/commercial/state"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/internal/commercial/wallets"
	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
)

// Module is the composition boundary for Leamout's Commercial domain.
// Runtime and telecom code should depend on this module rather than assembling
// Commercial subdomains independently.
type Module struct {
	Catalog  CatalogModule
	Access   AccessModule
	Metering MeteringModule
	Money    MoneyModule
	Payments PaymentsModule
	State    StateModule
}

type CatalogModule struct {
	Repository *catalog.Repository
	Service    *catalog.Service
	Handler    *catalog.Handler
}

type AccessModule struct {
	Subscriptions SubscriptionsModule
	Entitlements  EntitlementsModule
	Licensing     LicensingModule
}

type SubscriptionsModule struct {
	Repository *subscriptions.Repository
	Service    *subscriptions.Service
	Handler    *subscriptions.Handler
}

type EntitlementsModule struct {
	Repository *entitlements.Repository
	Service    *entitlements.Service
}

type LicensingModule struct {
	Repository *licensing.Repository
	Service    *licensing.Service
	Handler    *licensing.Handler
}

type MeteringModule struct {
	Repository *metering.Repository
	Service    *metering.Service
}

type MoneyModule struct {
	Wallets      *wallets.Repository
	TopupService *wallets.TopupService
	TopupHandler *wallets.TopupHandler
}

type PaymentsModule struct {
	Checkouts *checkout.Repository
	Payments  *payments.Repository
}

type StateModule struct {
	Service *commercialstate.Service
	Handler *commercialstate.Handler
}

// New composes Commercial from its durable subdomains. Payment providers are
// registered after construction so provider adapters remain outside Commercial
// state and can be configured by the runtime.
func New(db *pgxpool.Pool) *Module {
	catalogRepository := catalog.NewRepository(db)
	catalogService := catalog.NewService(catalogRepository)

	subscriptionsRepository := subscriptions.NewRepository(db)
	subscriptionsService := subscriptions.NewService(subscriptionsRepository, catalogService)

	entitlementsRepository := entitlements.NewRepository(db)
	entitlementsService := entitlements.NewService(entitlementsRepository, subscriptionsService)

	stateService := commercialstate.NewService(subscriptionsService, entitlementsService)

	licensingRepository := licensing.NewRepository(db)
	licensingService := licensing.NewService(licensingRepository, stateService)

	meteringRepository := metering.NewRepository(db)
	meteringService := metering.NewService(meteringRepository)

	walletRepository := wallets.NewRepository(db)
	checkoutRepository := checkout.NewRepository(db)
	paymentRepository := payments.NewRepository(db)
	topupService := wallets.NewTopupService(
		walletRepository,
		checkoutRepository,
		paymentRepository,
		walletRepository,
		map[string]paymentprovider.Provider{},
	)

	return &Module{
		Catalog: CatalogModule{
			Repository: catalogRepository,
			Service:    catalogService,
			Handler:    catalog.NewHandler(catalogService),
		},
		Access: AccessModule{
			Subscriptions: SubscriptionsModule{
				Repository: subscriptionsRepository,
				Service:    subscriptionsService,
				Handler:    subscriptions.NewHandler(subscriptionsService),
			},
			Entitlements: EntitlementsModule{
				Repository: entitlementsRepository,
				Service:    entitlementsService,
			},
			Licensing: LicensingModule{
				Repository: licensingRepository,
				Service:    licensingService,
				Handler:    licensing.NewHandler(licensingService),
			},
		},
		Metering: MeteringModule{
			Repository: meteringRepository,
			Service:    meteringService,
		},
		Money: MoneyModule{
			Wallets:      walletRepository,
			TopupService: topupService,
			TopupHandler: wallets.NewTopupHandler(topupService),
		},
		Payments: PaymentsModule{
			Checkouts: checkoutRepository,
			Payments:  paymentRepository,
		},
		State: StateModule{
			Service: stateService,
			Handler: commercialstate.NewHandler(stateService),
		},
	}
}
