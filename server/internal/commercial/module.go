package commercial

import (
	"github.com/leamout/leamout/server/internal/commercial/entitlements"
	"github.com/leamout/leamout/server/internal/commercial/plans"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
	"github.com/leamout/leamout/server/internal/database/sqlc"
)

type Module struct {
	Plans         PlansModule
	Subscriptions SubscriptionsModule
	Entitlements  EntitlementsModule
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

func New(queries *sqlc.Queries) *Module {
	plansRepository := plans.NewRepository(queries)
	plansService := plans.NewService(plansRepository)

	subscriptionsRepository := subscriptions.NewRepository(queries)
	subscriptionsService := subscriptions.NewService(subscriptionsRepository)

	entitlementsRepository := entitlements.NewRepository(queries)
	entitlementsService := entitlements.NewService(entitlementsRepository)

	return &Module{
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
