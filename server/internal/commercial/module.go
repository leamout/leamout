package commercial

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/commercial/catalog"
	checkout "github.com/leamout/leamout/internal/commercial/checkout"
	"github.com/leamout/leamout/internal/commercial/licensing"
	"github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/commercial/usage"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

type Module struct {
	Catalog   CatalogModule
	Billing   BillingModule
	Licensing LicensesModule
	Usage     UsageModule
	Wallets   WalletModule
}

type CatalogModule struct { Repository *catalog.Repository; Service *catalog.Service; Handler *catalog.Handler }
type BillingModule struct { Checkouts CheckoutModule; Payments PaymentsModule }
type CheckoutModule struct { Repository *checkout.Repository; Service *checkout.Service; Handler *checkout.Handler }
type LicensesModule struct { Repository *licensing.Repository; Service *licensing.Service; Handler *licensing.Handler }
type UsageModule struct { Repository *usage.Repository; Service *usage.Service }
type WalletModule struct { Repository *wallets.Repository; Service *wallets.Service; Handler *wallets.Handler }
type PaymentsModule struct { Repository *payments.Repository; Service *payments.Service; Providers *payments.ProviderRegistry; Handler *payments.Handler }

func New(db *pgxpool.Pool) *Module {
	catalogRepository := catalog.NewRepository(db)
	catalogService := catalog.NewService(catalogRepository)
	licensingRepository := licensing.NewRepository(db)
	licensingService := licensing.NewService(licensingRepository)
	usageRepository := usage.NewRepository(db)
	usageService := usage.NewService(usageRepository)
	walletRepository := wallets.NewRepository(db)
	walletService := wallets.NewService(walletRepository, catalogService)
	paymentRepository := payments.NewRepository(db)
	providerRegistry := payments.NewProviderRegistry()
	paymentService := payments.NewService(paymentRepository, providerRegistry)
	checkoutRepository := checkout.NewRepository(db)
	checkoutService := checkout.NewService(checkoutRepository, walletService, paymentService)
	return &Module{
		Catalog: CatalogModule{Repository: catalogRepository, Service: catalogService, Handler: catalog.NewHandler(catalogService)},
		Billing: BillingModule{
			Checkouts: CheckoutModule{Repository: checkoutRepository, Service: checkoutService, Handler: checkout.NewHandler(checkoutService)},
			Payments: PaymentsModule{Repository: paymentRepository, Service: paymentService, Providers: providerRegistry, Handler: payments.NewHandler(paymentService, checkoutService)},
		},
		Licensing: LicensesModule{Repository: licensingRepository, Service: licensingService, Handler: licensing.NewHandler(licensingService)},
		Usage: UsageModule{Repository: usageRepository, Service: usageService},
		Wallets: WalletModule{Repository: walletRepository, Service: walletService, Handler: wallets.NewHandler(walletService)},
	}
}
