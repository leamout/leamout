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

// Module owns the Commercial services used by the application.
type Module struct {
	Catalog       *catalog.Service
	Subscriptions *subscriptions.Service
	Entitlements  *entitlements.Service
	Licensing     *licensing.Service
	Metering      *metering.Service
	Wallets       *wallets.Repository
	Topups        *wallets.TopupService
	State         *commercialstate.Service

	handlers handlers
}

type handlers struct {
	catalog       *catalog.Handler
	subscriptions *subscriptions.Handler
	licensing     *licensing.Handler
	topups        *wallets.TopupHandler
	state         *commercialstate.Handler
}

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
	topupRepository := wallets.NewTopupRepository(db)
	topupService := wallets.NewTopupService(
		walletRepository,
		checkoutRepository,
		paymentRepository,
		topupRepository,
		map[string]paymentprovider.Provider{},
	)

	return &Module{
		Catalog:       catalogService,
		Subscriptions: subscriptionsService,
		Entitlements:  entitlementsService,
		Licensing:     licensingService,
		Metering:      meteringService,
		Wallets:       walletRepository,
		Topups:        topupService,
		State:         stateService,
		handlers: handlers{
			catalog:       catalog.NewHandler(catalogService),
			subscriptions: subscriptions.NewHandler(subscriptionsService),
			licensing:     licensing.NewHandler(licensingService),
			topups:        wallets.NewTopupHandler(topupService),
			state:         commercialstate.NewHandler(stateService),
		},
	}
}
