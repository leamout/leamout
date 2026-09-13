package worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/leamout/leamout/internal/integrations/freeswitch"
)

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
