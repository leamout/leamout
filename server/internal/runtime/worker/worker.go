package worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/integrations/carriers/didww"
	"github.com/leamout/leamout/internal/integrations/freeswitch"
	natsintegration "github.com/leamout/leamout/internal/integrations/nats"
	redisintegration "github.com/leamout/leamout/internal/integrations/redis"
	"github.com/leamout/leamout/internal/modules/idempotency"
	"github.com/leamout/leamout/internal/modules/outbox"
	"github.com/leamout/leamout/internal/modules/webhooks"
	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/platform/logging"
	"github.com/leamout/leamout/internal/platform/metrics"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/numbers"
	"github.com/leamout/leamout/internal/telecom/recordings"
	"github.com/leamout/leamout/internal/telecom/routing"
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
	outbox                  *outbox.PublisherJob
	webhookConsumer         *webhooks.Consumer
	webhookDelivery         *webhooks.DeliveryJob
	idempotencyCleanup      *idempotency.CleanupJob
	health                  *healthState
	logger                  *logging.Logger
}

const workerHealthAddress = ":8081"

var componentNames = []string{
	"freeswitch-events",
	"call-reconciliation",
	"carrier-endpoint-health",
	"recording-reconciliation",
	"provider-operations",
	"outbox-publisher",
	"webhook-consumer",
	"webhook-delivery",
	"idempotency-cleanup",
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
	redisClient, err := redisintegration.New(ctx, redisintegration.DefaultConfig(cfg.RedisURL))
	if err != nil {
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("connect worker Redis: %w", err)
	}
	streamLimits := natsintegration.DefaultStreamLimits()
	streamLimits.Replicas = cfg.NATSStreamReplicas
	if streamLimits.Replicas <= 0 {
		streamLimits.Replicas = 1
	}
	if err := natsClient.Provision(ctx, streamLimits); err != nil {
		_ = redisClient.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("provision worker NATS streams: %w", err)
	}

	freeSwitch, err := freeswitch.New(freeswitch.DefaultConfig(cfg.FreeSWITCHESLAddress, cfg.FreeSWITCHESLPassword))
	if err != nil {
		_ = redisClient.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize worker FreeSWITCH client: %w", err)
	}
	if err := freeSwitch.Connect(ctx); err != nil {
		_ = redisClient.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("connect worker FreeSWITCH: %w", err)
	}

	queries := sqlc.New(db)
	routingRepository := routing.NewRepository(queries)
	routeResolver := routing.NewResolver(routingRepository)
	telecomMetrics := metrics.New(redisClient)
	routeResolver.SetMetrics(telecomMetrics)
	routingService := routing.NewService(routeResolver)
	callsRepository := calls.NewRepository(db)
	controller := calls.NewFreeSWITCHController(freeSwitch)
	callAdmission, err := calls.NewAdmissionController(redisClient, callsRepository)
	if err != nil {
		_ = redisClient.Close()
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize call admission: %w", err)
	}
	callsService := calls.NewService(callsRepository, controller, routingService, callAdmission)
	callsService.SetMetrics(telecomMetrics)

	recordingsRepository := recordings.NewRepository(db)
	recordingsService := recordings.NewService(recordingsRepository, nil)

	callReconciliation, err := calls.NewReconciliationJob(
		callsRepository,
		freeSwitch,
		calls.DefaultReconciliationJobConfig(),
		callAdmission,
	)
	if err != nil {
		_ = redisClient.Close()
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize call reconciliation job: %w", err)
	}
	callReconciliation.SetMetrics(telecomMetrics)
	endpointHealth, err := routing.NewEndpointHealthJob(queries, routing.NewSIPOptionsProber())
	if err != nil {
		_ = redisClient.Close()
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize carrier endpoint health job: %w", err)
	}
	endpointHealth.SetMetrics(telecomMetrics)

	recordingReconciliation, err := recordings.NewReconciliationJob(recordingsRepository, recordings.DefaultReconciliationJobConfig())
	if err != nil {
		_ = redisClient.Close()
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize recording reconciliation job: %w", err)
	}

	numbersRepository := numbers.NewRepository(db, redisClient)
	numbersService := numbers.NewService(numbersRepository)
	if strings.TrimSpace(cfg.DIDWW.APIKey) != "" {
		didwwClient, err := didww.NewClient(didww.Config{BaseURL: cfg.DIDWW.APIBaseURL, APIKey: cfg.DIDWW.APIKey})
		if err != nil {
			_ = redisClient.Close()
			_ = freeSwitch.Close()
			_ = natsClient.Close()
			db.Close()
			return nil, fmt.Errorf("initialize DIDWW provider executor: %w", err)
		}
		numbersService.SetManagedProvider("didww", didwwClient)
	}
	providerOperations, err := numbers.NewProviderOperationJob(
		numbersRepository,
		numbersService,
		numbers.DefaultProviderOperationJobConfig(),
	)
	if err != nil {
		_ = redisClient.Close()
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize provider operation job: %w", err)
	}

	outboxRepository := outbox.NewRepository(queries)
	outboxPublisher := outbox.NewPublisher(natsClient)
	outboxJob, err := outbox.NewPublisherJob(outboxRepository, outboxPublisher, outbox.DefaultPublisherJobConfig("worker-"+uuid.NewString()))
	if err != nil {
		_ = redisClient.Close()
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize outbox publisher job: %w", err)
	}

	webhookRepository := webhooks.NewRepository(queries)
	webhookService := webhooks.NewService(webhookRepository, db)
	webhookConsumer := webhooks.NewConsumer(natsClient, webhookService)
	webhookDelivery, err := webhooks.NewDeliveryJob(webhookRepository, webhooks.NewHTTPSender(), webhooks.DefaultDeliveryJobConfig("worker-"+uuid.NewString()))
	if err != nil {
		_ = redisClient.Close()
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize webhook delivery job: %w", err)
	}
	idempotencyCleanup, err := idempotency.NewCleanupJob(idempotency.NewRepository(queries), time.Hour)
	if err != nil {
		_ = redisClient.Close()
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize idempotency cleanup job: %w", err)
	}

	return &Worker{
		db: db,
		freeSwitch: freeSwitch,
		nats: natsClient,
		redis: redisClient,
		calls: calls.NewConsumer(callsService),
		recordings: recordings.NewConsumer(recordingsService),
		callReconciliation: callReconciliation,
		endpointHealth: endpointHealth,
		recordingReconciliation: recordingReconciliation,
		providerOperations: providerOperations,
		outbox: outboxJob,
		webhookConsumer: webhookConsumer,
		webhookDelivery: webhookDelivery,
		idempotencyCleanup: idempotencyCleanup,
		health: newHealthState(componentNames...),
		logger: logging.New(),
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
		Addr: workerHealthAddress,
		Handler: healthHandler(w.db, w.nats, w.redis, w.freeSwitch, w.health),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout: 30 * time.Second,
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = healthServer.Shutdown(shutdownCtx)
	}()

	errCh := make(chan error, len(componentNames)+1)
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

func (w *Worker) runComponent(ctx context.Context, errCh chan<- error, name string, run func(context.Context) error) {
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
