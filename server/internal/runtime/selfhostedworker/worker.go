package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/database/sqlc"
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
	outbox                  *outbox.PublisherJob
	webhookConsumer         *webhooks.Consumer
	webhookDelivery         *webhooks.DeliveryJob
	idempotencyCleanup      *idempotency.CleanupJob
	health                  *healthState
	logger                  *logging.Logger
	componentNames          []string
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
		_ = freeSwitch.Close()
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

	recordingReconciliation, err := recordings.NewReconciliationJob(
		recordingsRepository,
		recordings.DefaultReconciliationJobConfig(),
	)
	if err != nil {
		_ = redisClient.Close()
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
		_ = redisClient.Close()
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
		_ = redisClient.Close()
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize webhook delivery job: %w", err)
	}

	idempotencyCleanup, err := idempotency.NewCleanupJob(
		idempotency.NewRepository(queries),
		time.Hour,
	)
	if err != nil {
		_ = redisClient.Close()
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize idempotency cleanup job: %w", err)
	}

	components := append([]string(nil), componentNames...)
	return &Worker{
		db:                      db,
		freeSwitch:              freeSwitch,
		nats:                    natsClient,
		redis:                   redisClient,
		calls:                   calls.NewConsumer(callsService),
		recordings:              recordings.NewConsumer(recordingsService),
		callReconciliation:      callReconciliation,
		endpointHealth:          endpointHealth,
		recordingReconciliation: recordingReconciliation,
		outbox:                  outboxJob,
		webhookConsumer:         webhookConsumer,
		webhookDelivery:         webhookDelivery,
		idempotencyCleanup:      idempotencyCleanup,
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
