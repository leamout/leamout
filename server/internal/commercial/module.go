package commercial

import (
	"github.com/leamout/leamout/server/internal/commercial/billing"
	"github.com/leamout/leamout/server/internal/commercial/entitlements"
	"github.com/leamout/leamout/server/internal/commercial/licensing"
	"github.com/leamout/leamout/server/internal/commercial/limits"
	"github.com/leamout/leamout/server/internal/commercial/plans"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	polarintegration "github.com/leamout/leamout/server/internal/integrations/payments/polar"
)

type Module struct {
	License       *licensing.Service
	Billing       BillingModule
	Plans         PlansModule
	Subscriptions SubscriptionsModule
	Limits        LimitsModule
	Entitlements  EntitlementsModule
}

type BillingModule struct {
	Service *billing.Service
	Handler *billing.Handler
}

type Dependencies struct {
	License       *licensing.Service
	Polar         *polarintegration.Client
	PolarProducts map[string]string
	FrontendURL   string
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

type LimitsModule struct {
	Service *limits.Service
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
	limitsService := limits.NewService(plansService, subscriptionsService, queries)

	limitsService.SetLicenseService(deps.License)

	billingService := billing.NewService(
		plansService,
		subscriptionsService,
		deps.Polar,
		deps.PolarProducts,
		deps.FrontendURL,
	)

	entitlementsRepository := entitlements.NewRepository(queries)
	entitlementsService := entitlements.NewService(entitlementsRepository)
	entitlementsService.SetLicenseService(deps.License)

	return &Module{
		License: deps.License,
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
		Limits: LimitsModule{
			Service: limitsService,
		},
		Entitlements: EntitlementsModule{
			Repository: entitlementsRepository,
			Service:    entitlementsService,
			Handler:    entitlements.NewHandler(entitlementsService),
			Middleware: entitlements.NewMiddleware(entitlementsService),
		},
	}
}
