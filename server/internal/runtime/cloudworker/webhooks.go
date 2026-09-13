package worker

import (
	"github.com/google/uuid"

	"github.com/leamout/leamout/internal/modules/webhooks"
)

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
