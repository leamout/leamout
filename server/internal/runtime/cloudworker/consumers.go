package worker

import (
	"github.com/google/uuid"

	"github.com/leamout/leamout/internal/modules/webhooks"
	"github.com/leamout/leamout/internal/platform/metrics"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/recordings"
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

type recordingJobs struct {
	consumer       *recordings.Consumer
	reconciliation *recordings.ReconciliationJob
}

func newRecordingJobs(deps *dependencies) (*recordingJobs, error) {
	repository := recordings.NewRepository(deps.db)
	service := recordings.NewService(repository, nil)
	reconciliation, err := recordings.NewReconciliationJob(
		repository,
		recordings.DefaultReconciliationJobConfig(),
	)
	if err != nil {
		return nil, wrapWorkerError("initialize recording reconciliation job", err)
	}
	return &recordingJobs{
		consumer:       recordings.NewConsumer(service),
		reconciliation: reconciliation,
	}, nil
}

type webhookJobs struct {
	consumer *webhooks.Consumer
	delivery *webhooks.DeliveryJob
}

func newWebhookJobs(deps *dependencies) (*webhookJobs, error) {
	repository := webhooks.NewRepository(deps.queries)
	service := webhooks.NewService(repository, deps.db)
	consumer := webhooks.NewConsumer(deps.nats, service)
	delivery, err := webhooks.NewDeliveryJob(
		repository,
		webhooks.NewHTTPSender(),
		webhooks.DefaultDeliveryJobConfig("worker-"+uuid.NewString()),
	)
	if err != nil {
		return nil, wrapWorkerError("initialize webhook delivery job", err)
	}
	return &webhookJobs{consumer: consumer, delivery: delivery}, nil
}
