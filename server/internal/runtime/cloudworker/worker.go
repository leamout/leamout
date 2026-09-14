package worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/integrations/freeswitch"
	natsintegration "github.com/leamout/leamout/internal/integrations/nats"
	redisintegration "github.com/leamout/leamout/internal/integrations/redis"
	"github.com/leamout/leamout/internal/modules/idempotency"
	"github.com/leamout/leamout/internal/modules/outbox"
	"github.com/leamout/leamout/internal/modules/webhooks"
	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/platform/logging"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/numbers"
	"github.com/leamout/leamout/internal/telecom/recordings"
	"github.com/leamout/leamout/internal/telecom/routing"
	"github.com/leamout/leamout/internal/telecom/wholesale"
)

type Worker struct {
	db                      *pgxpool.Pool
	freeSwitch              *freeswitch.Client
	nats                    *natsintegration.Client
	redis                   *redisintegration.Client
	calls                   *calls.Consumer
	recordings              *recordings.Consumer
	callReconciliation      *calls.ReconciliationJob
	endpointHealth          *routing.EndpointHealthJob
	recordingReconciliation *recordings.ReconciliationJob
	providerOperations      *numbers.ProviderOperationJob
	commpeakCDRPolling      *wholesale.CDRPollJob
	outbox                  *outbox.PublisherJob
	webhookConsumer         *webhooks.Consumer
	webhookDelivery         *webhooks.DeliveryJob
	idempotencyCleanup      *idempotency.CleanupJob
	health                  *healthState
	logger                  *logging.Logger
	componentNames          []string
}

const workerHealthAddress = ":8081"

var componentNames = []string{
	"freeswitch-events",
	"call-reconciliation",
	"carrier-endpoint-health",
	"recording-reconciliation",
	"provider-operations",
	"commpeak-cdr-polling",
	"outbox-publisher",
	"webhook-consumer",
	"webhook-delivery",
	"idempotency-cleanup",
}

func New(ctx context.Context, cfg config.CloudConfig) (*Worker, error) {
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

func (w *Worker) Run(ctx context.Context) error {
	events := []string{
		"CHANNEL_CREATE",
		"CHANNEL_ANSWER",
		"CHANNEL_HOLD",
		"CHANNEL_UNHOLD",
		"CHANNEL_HANGUP_COMPLETE",
		"RECORD_START",
		"RECORD_STOP",
	}

	if err := w.freeSwitch.Subscribe(
		ctx,
		freeswitch.EventFormatPlain,
		events,
		func(eventCtx context.Context, event freeswitch.Event) error {
			if err := w.calls.HandleFreeSWITCHEvent(eventCtx, event); err != nil {
				w.logger.Error(eventCtx, "FreeSWITCH call event failed", "event", event.Name, "error", err)
				return err
			}
			if err := w.recordings.HandleFreeSWITCHEvent(eventCtx, event); err != nil {
				w.logger.Error(eventCtx, "FreeSWITCH recording event failed", "event", event.Name, "error", err)
				return err
			}
			return nil
		},
	); err != nil {
		return fmt.Errorf("subscribe to FreeSWITCH lifecycle events: %w", err)
	}
	w.health.setRunning("freeswitch-events")
	w.logger.Info(ctx, "worker subscribed to FreeSWITCH lifecycle events")

	healthServer := &http.Server{
		Addr:              workerHealthAddress,
		Handler:           healthHandler(w.db, w.nats, w.redis, w.freeSwitch, w.health),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = healthServer.Shutdown(shutdownCtx)
	}()

	errCh := make(chan error, len(w.componentNames)+1)
	go func() {
		w.logger.Info(ctx, "worker health server started", "address", workerHealthAddress)
		if err := healthServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("run health server: %w", err)
		}
	}()
	go w.runComponent(ctx, errCh, "call-reconciliation", w.callReconciliation.Run)
	go w.runComponent(ctx, errCh, "carrier-endpoint-health", w.endpointHealth.Run)
	go w.runComponent(ctx, errCh, "recording-reconciliation", w.recordingReconciliation.Run)
	go w.runComponent(ctx, errCh, "provider-operations", w.providerOperations.Run)
	go w.runComponent(ctx, errCh, "commpeak-cdr-polling", w.commpeakCDRPolling.Run)
	go w.runComponent(ctx, errCh, "outbox-publisher", w.outbox.Run)
	go w.runComponent(ctx, errCh, "webhook-consumer", w.webhookConsumer.Run)
	go w.runComponent(ctx, errCh, "webhook-delivery", w.webhookDelivery.Run)
	go w.runComponent(ctx, errCh, "idempotency-cleanup", w.idempotencyCleanup.Run)

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (w *Worker) runComponent(
	ctx context.Context,
	errCh chan<- error,
	name string,
	run func(context.Context) error,
) {
	w.health.setRunning(name)
	w.logger.Info(ctx, "worker component started", "component", name)
	err := run(ctx)
	if ctx.Err() != nil {
		w.health.setStopped(name, nil)
		return
	}
	if err == nil {
		err = fmt.Errorf("component stopped unexpectedly")
	}
	w.health.setStopped(name, err)
	w.logger.Error(ctx, "worker component stopped", "component", name, "error", err)
	errCh <- fmt.Errorf("run %s: %w", name, err)
}
