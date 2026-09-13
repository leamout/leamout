package worker

import (
	"github.com/leamout/leamout/internal/platform/metrics"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/routing"
)

type callJobs struct {
	consumer       *calls.Consumer
	reconciliation *calls.ReconciliationJob
	endpointHealth *routing.EndpointHealthJob
}

func newCallJobs(deps *dependencies) (*callJobs, error) {
	routingRepository := routing.NewRepository(deps.queries)
	routeResolver := routing.NewResolver(routingRepository)
	telecomMetrics := metrics.New(deps.redis)
	routeResolver.SetMetrics(telecomMetrics)
	routingService := routing.NewService(routeResolver)
	callsRepository := calls.NewRepository(deps.db)
	controller := calls.NewFreeSWITCHController(deps.freeSwitch)
	callAdmission, err := calls.NewAdmissionController(deps.redis, callsRepository)
	if err != nil {
		return nil, wrapWorkerError("initialize call admission", err)
	}
	callsService := calls.NewService(callsRepository, controller, routingService, callAdmission)
	callsService.SetMetrics(telecomMetrics)

	reconciliation, err := calls.NewReconciliationJob(
		callsRepository,
		deps.freeSwitch,
		calls.DefaultReconciliationJobConfig(),
		callAdmission,
	)
	if err != nil {
		return nil, wrapWorkerError("initialize call reconciliation job", err)
	}
	reconciliation.SetMetrics(telecomMetrics)

	endpointHealth, err := routing.NewEndpointHealthJob(deps.queries, routing.NewSIPOptionsProber())
	if err != nil {
		return nil, wrapWorkerError("initialize carrier endpoint health job", err)
	}
	endpointHealth.SetMetrics(telecomMetrics)

	return &callJobs{
		consumer:       calls.NewConsumer(callsService),
		reconciliation: reconciliation,
		endpointHealth: endpointHealth,
	}, nil
}
