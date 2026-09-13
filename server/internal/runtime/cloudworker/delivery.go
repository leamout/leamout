package worker

import (
	"github.com/leamout/leamout/internal/modules/idempotency"
	"github.com/leamout/leamout/internal/modules/outbox"
	"github.com/leamout/leamout/internal/modules/webhooks"
)

type deliveryJobs struct {
	outbox             *outbox.PublisherJob
	webhookConsumer    *webhooks.Consumer
	webhookDelivery    *webhooks.DeliveryJob
	idempotencyCleanup *idempotency.CleanupJob
}

func newDeliveryJobs(deps *dependencies) (*deliveryJobs, error) {
	outboxJob, err := newOutboxJob(deps)
	if err != nil {
		return nil, err
	}
	webhookJobs, err := newWebhookJobs(deps)
	if err != nil {
		return nil, err
	}
	cleanup, err := newIdempotencyCleanup(deps)
	if err != nil {
		return nil, err
	}
	return &deliveryJobs{
		outbox:             outboxJob,
		webhookConsumer:    webhookJobs.consumer,
		webhookDelivery:    webhookJobs.delivery,
		idempotencyCleanup: cleanup,
	}, nil
}
