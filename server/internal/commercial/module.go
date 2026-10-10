package commercial

import (
	"github.com/leamout/leamout/server/internal/commercial/billing"
	"github.com/leamout/leamout/server/internal/commercial/entitlements"
	"github.com/leamout/leamout/server/internal/commercial/plans"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	stripeintegration "github.com/leamout/leamout/server/internal/integrations/payments/stripe"
)

type Module struct {
	Billing       BillingModule
	Plans         PlansModule
	Subscriptions SubscriptionsModule
	Entitlements  EntitlementsModule
}

type BillingModule struct {
	Service *billing.Service
	Handler *billing.Handler
}

type Dependencies struct {
	Stripe       *stripeintegration.Client
	StripePrices map[string]string
}

type PlansModule struct {
	Repository *plans.Repository
	Service    *plans.Service
	Handler    *plans.Handler
}

type SubscriptionsModule struct {
	Repository *subscriptions.Repository
	Service    *subscriptions.Service
	Handler    *subscriptions.Handler
}

type EntitlementsModule struct {
	Repository *entitlements.Repository
	Service    *entitlements.Service
	Handler    *entitlements.Handler
	Middleware *entitlements.Middleware
}

func New(queries *sqlc.Queries, dependencies ...Dependencies) *Module {
	var deps Dependencies
	if len(dependencies) > 0 {
		deps = dependencies[0]
	}
	plansRepository := plans.NewRepository(queries)
	plansService := plans.NewService(plansRepository)

	subscriptionsRepository := subscriptions.NewRepository(queries)
	subscriptionsService := subscriptions.NewService(subscriptionsRepository)

	billingService := billing.NewService(
		plansService,
		subscriptionsService,
		deps.Stripe,
		deps.StripePrices,
	)

	entitlementsRepository := entitlements.NewRepository(queries)
	entitlementsService := entitlements.NewService(entitlementsRepository)

	return &Module{
		Billing: BillingModule{
			Service: billingService,
			Handler: billing.NewHandler(billingService),
		},
		Plans: PlansModule{
			Repository: plansRepository,
			Service:    plansService,
			Handler:    plans.NewHandler(plansService),
		},
		Subscriptions: SubscriptionsModule{
			Repository: subscriptionsRepository,
			Service:    subscriptionsService,
			Handler:    subscriptions.NewHandler(subscriptionsService),
		},
		Entitlements: EntitlementsModule{
			Repository: entitlementsRepository,
			Service:    entitlementsService,
			Handler:    entitlements.NewHandler(entitlementsService),
			Middleware: entitlements.NewMiddleware(entitlementsService),
		},
	}
}
