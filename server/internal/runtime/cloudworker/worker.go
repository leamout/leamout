package worker

import (
	"context"
	"fmt"

	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/platform/logging"
)

func New(ctx context.Context, cfg config.Config) (*Worker, error) {
	deps, err := newDependencies(ctx, cfg)
	if err != nil {
		return nil, err
	}

	shared, err := newSharedJobs(deps)
	if err != nil {
		deps.close()
		return nil, err
	}
	providerOperations, commpeakCDRPolling, err := newManagedJobs(deps.db, deps.redis, cfg)
	if err != nil {
		deps.close()
		return nil, err
	}
	delivery, err := newDeliveryJobs(deps)
	if err != nil {
		deps.close()
		return nil, err
	}

	components := append([]string(nil), componentNames...)
	return &Worker{
		db:                      deps.db,
		freeSwitch:              deps.freeSwitch,
		nats:                    deps.nats,
		redis:                   deps.redis,
		calls:                   shared.calls.consumer,
		recordings:              shared.recordings.consumer,
		callReconciliation:      shared.calls.reconciliation,
		endpointHealth:          shared.calls.endpointHealth,
		recordingReconciliation: shared.recordings.reconciliation,
		providerOperations:      providerOperations,
		commpeakCDRPolling:      commpeakCDRPolling,
		outbox:                  delivery.outbox,
		webhookConsumer:         delivery.webhookConsumer,
		webhookDelivery:         delivery.webhookDelivery,
		idempotencyCleanup:      delivery.idempotencyCleanup,
		health:                  newHealthState(components...),
		logger:                  logging.New(),
		componentNames:          components,
	}, nil
}

func (w *Worker) Close() {
	if w.freeSwitch != nil {
		_ = w.freeSwitch.Close()
	}
	if w.nats != nil {
		_ = w.nats.Close()
	}
	if w.redis != nil {
		_ = w.redis.Close()
	}
	if w.db != nil {
		w.db.Close()
	}
}

func wrapWorkerError(operation string, err error) error {
	return fmt.Errorf("%s: %w", operation, err)
}
