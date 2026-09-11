package commercial

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/commercial/catalog"
	checkout "github.com/leamout/leamout/internal/commercial/checkout"
	"github.com/leamout/leamout/internal/commercial/entitlements"
	"github.com/leamout/leamout/internal/commercial/licensing"
	"github.com/leamout/leamout/internal/commercial/orders"
	"github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/commercial/purchase"
	commercialstate "github.com/leamout/leamout/internal/commercial/state"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/internal/commercial/usage"
	"github.com/leamout/leamout/internal/commercial/wallets"
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
}

type CatalogModule struct {
	Repository *catalog.Repository
	Service    *catalog.Service
	Handler    *catalog.Handler
}

type PurchaseModule struct {
	Checkouts CheckoutModule
	Orders    *orders.Repository
	Service   *purchase.Service
}

type CheckoutModule struct {
	Repository *checkout.Repository
	Service    *checkout.Service
	Handler    *checkout.Handler
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
	Wallets WalletModule
}

type WalletModule struct {
	Repository *wallets.Repository
	Service    *wallets.Service
	Handler    *wallets.Handler
}

type PaymentsModule struct {
	Repository *payments.Repository
	Service    *payments.Service
	Providers  *payments.ProviderRegistry
	Handler    *payments.Handler
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
	walletService := wallets.NewService(walletRepository)
	checkoutRepository := checkout.NewRepository(db)
	checkoutService := checkout.NewService(checkoutRepository)
	orderRepository := orders.NewRepository(db)
	purchaseService := purchase.NewService()
	paymentRepository := payments.NewRepository(db, purchaseService)
	paymentService := payments.NewService(paymentRepository)
	providerRegistry := payments.NewProviderRegistry()
	topupService := checkout.NewTopupService(
		walletService,
		checkoutService,
		paymentRepository,
		paymentService,
		providerRegistry,
	)
	checkoutHandler := checkout.NewHandler(topupService)
	walletHandler := wallets.NewHandler(walletTopupCreator(topupService))
	stateModule := StateModule{
		Service: commercialStateService,
		Handler: commercialstate.NewHandler(commercialStateService),
	}
	prepaidModule := PrepaidModule{
		Wallets: WalletModule{Repository: walletRepository, Service: walletService, Handler: walletHandler},
	}

	return &Module{
		Catalog: CatalogModule{
			Repository: catalogRepository,
			Service:    catalogService,
			Handler:    catalog.NewHandler(catalogService),
		},
		Purchase: PurchaseModule{
			Checkouts: CheckoutModule{Repository: checkoutRepository, Service: checkoutService, Handler: checkoutHandler},
			Orders:    orderRepository,
			Service:   purchaseService,
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
			Service:    paymentService,
			Providers:  providerRegistry,
			Handler:    payments.NewHandler(paymentService, providerRegistry),
		},
	}
}

func walletTopupCreator(service *checkout.TopupService) wallets.TopupCreator {
	return func(ctx context.Context, organizationID, walletID uuid.UUID, input wallets.TopupRequest) (wallets.TopupResponse, error) {
		var mobileMoney *payments.MobileMoney
		if input.MobileMoney != nil {
			mobileMoney = &payments.MobileMoney{Phone: input.MobileMoney.Phone, Provider: input.MobileMoney.Provider}
		}
		result, err := service.Create(ctx, organizationID, walletID, checkout.TopupCreateInput{
			AmountMinor: input.AmountMinor, Provider: checkout.Provider(input.Provider), Email: input.Email,
			CallbackURL: input.CallbackURL, MobileMoney: mobileMoney,
		})
		if err != nil {
			return wallets.TopupResponse{}, err
		}
		response := checkout.Response(result)
		return wallets.TopupResponse{
			CheckoutID: response.CheckoutID, PaymentID: response.PaymentID, Reference: response.Reference,
			Provider: string(response.Provider), AmountMinor: response.AmountMinor, Currency: response.Currency,
			Status: string(response.Status), NextAction: string(response.NextAction),
			ProviderMessage: response.ProviderMessage, ClientSecret: response.ClientSecret,
		}, nil
	}
}
