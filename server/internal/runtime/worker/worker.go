package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/integrations/freeswitch"
	natsintegration "github.com/leamout/leamout/internal/integrations/nats"
	"github.com/leamout/leamout/internal/modules/outbox"
	"github.com/leamout/leamout/internal/modules/webhooks"
	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/recordings"
	"github.com/leamout/leamout/internal/telecom/routing"
)

type Worker struct {
	db                      *pgxpool.Pool
	freeSwitch              *freeswitch.Client
	nats                    *natsintegration.Client
	calls                   *calls.Consumer
	recordings              *recordings.Consumer
	callReconciliation      *calls.ReconciliationJob
	recordingReconciliation *recordings.ReconciliationJob
	outbox                  *outbox.PublisherJob
	webhookConsumer         *webhooks.Consumer
	webhookDelivery         *webhooks.DeliveryJob
}

func New(ctx context.Context, cfg config.Config) (*Worker, error) {
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect worker database: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping worker database: %w", err)
	}

	natsClient, err := natsintegration.New(ctx, natsintegration.DefaultConfig(cfg.NATSURL))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("connect worker NATS: %w", err)
	}
	streamLimits := natsintegration.DefaultStreamLimits()
	streamLimits.Replicas = cfg.NATSStreamReplicas
	if streamLimits.Replicas <= 0 {
		streamLimits.Replicas = 1
	}
	if err := natsClient.Provision(ctx, streamLimits); err != nil {
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("provision worker NATS streams: %w", err)
	}

	freeSwitch, err := freeswitch.New(
		freeswitch.DefaultConfig(cfg.FreeSWITCHESLAddress, cfg.FreeSWITCHESLPassword),
	)
	if err != nil {
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize worker FreeSWITCH client: %w", err)
	}
	if err := freeSwitch.Connect(ctx); err != nil {
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("connect worker FreeSWITCH: %w", err)
	}

	queries := sqlc.New(db)
	routingRepository := routing.NewRepository(queries)
	routeResolver := routing.NewResolver(routingRepository)
	routingService := routing.NewService(routeResolver)
	callsRepository := calls.NewRepository(db)
	controller := calls.NewFreeSWITCHController(freeSwitch)
	callsService := calls.NewService(callsRepository, controller, routingService)

	recordingsRepository := recordings.NewRepository(db)
	recordingsService := recordings.NewService(recordingsRepository, nil)

	callReconciliation, err := calls.NewReconciliationJob(
		callsRepository,
		freeSwitch,
		calls.DefaultReconciliationJobConfig(),
	)
	if err != nil {
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize call reconciliation job: %w", err)
	}

	recordingReconciliation, err := recordings.NewReconciliationJob(
		recordingsRepository,
		recordings.DefaultReconciliationJobConfig(),
	)
	if err != nil {
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize recording reconciliation job: %w", err)
	}

	outboxRepository := outbox.NewRepository(queries)
	outboxPublisher := outbox.NewPublisher(natsClient)
	outboxJob, err := outbox.NewPublisherJob(
		outboxRepository,
		outboxPublisher,
		outbox.DefaultPublisherJobConfig("worker-"+uuid.NewString()),
	)
	if err != nil {
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize outbox publisher job: %w", err)
	}

	webhookRepository := webhooks.NewRepository(queries)
	webhookService := webhooks.NewService(webhookRepository, db)
	webhookConsumer := webhooks.NewConsumer(natsClient, webhookService)
	webhookDelivery, err := webhooks.NewDeliveryJob(
		webhookRepository,
		webhooks.NewHTTPSender(),
		webhooks.DefaultDeliveryJobConfig("worker-"+uuid.NewString()),
	)
	if err != nil {
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize webhook delivery job: %w", err)
	}

	return &Worker{
		db:                      db,
		freeSwitch:              freeSwitch,
		nats:                    natsClient,
		calls:                   calls.NewConsumer(callsService),
		recordings:              recordings.NewConsumer(recordingsService),
		callReconciliation:      callReconciliation,
		recordingReconciliation: recordingReconciliation,
		outbox:                  outboxJob,
		webhookConsumer:         webhookConsumer,
		webhookDelivery:         webhookDelivery,
	}, nil
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
				log.Printf("FreeSWITCH call event %s failed: %v", event.Name, err)
				return err
			}
			if err := w.recordings.HandleFreeSWITCHEvent(eventCtx, event); err != nil {
				log.Printf("FreeSWITCH recording event %s failed: %v", event.Name, err)
				return err
			}
			return nil
		},
	); err != nil {
		return fmt.Errorf("subscribe to FreeSWITCH lifecycle events: %w", err)
	}

	log.Print("worker subscribed to FreeSWITCH call and recording lifecycle events")
	log.Print("worker started call reconciliation job")
	log.Print("worker started recording reconciliation job")
	log.Print("worker started outbox NATS publisher")
	log.Print("worker started webhook NATS consumer")
	log.Print("worker started webhook delivery job")

	errCh := make(chan error, 5)
	go runComponent(ctx, errCh, "call reconciliation", w.callReconciliation.Run)
	go runComponent(ctx, errCh, "recording reconciliation", w.recordingReconciliation.Run)
	go runComponent(ctx, errCh, "outbox publisher", w.outbox.Run)
	go runComponent(ctx, errCh, "webhook consumer", w.webhookConsumer.Run)
	go runComponent(ctx, errCh, "webhook delivery", w.webhookDelivery.Run)

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func runComponent(ctx context.Context, errCh chan<- error, name string, run func(context.Context) error) {
	if err := run(ctx); err != nil {
		errCh <- fmt.Errorf("run %s: %w", name, err)
	}
}

func (w *Worker) Close() {
	if w.freeSwitch != nil {
		_ = w.freeSwitch.Close()
	}
	if w.nats != nil {
		_ = w.nats.Close()
	}
	if w.db != nil {
		w.db.Close()
	}
}
