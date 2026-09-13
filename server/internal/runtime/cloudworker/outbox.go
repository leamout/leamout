package worker

import (
	"github.com/google/uuid"

	"github.com/leamout/leamout/internal/modules/outbox"
)

func newOutboxJob(deps *dependencies) (*outbox.PublisherJob, error) {
	repository := outbox.NewRepository(deps.queries)
	publisher := outbox.NewPublisher(deps.nats)
	job, err := outbox.NewPublisherJob(
		repository,
		publisher,
		outbox.DefaultPublisherJobConfig("worker-"+uuid.NewString()),
	)
	if err != nil {
		return nil, wrapWorkerError("initialize outbox publisher job", err)
	}
	return job, nil
}
