package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/leamout/leamout/server/internal/commercial"
	"github.com/leamout/leamout/server/internal/commercial/billing"
	"github.com/leamout/leamout/server/internal/integrations/payments/polar"

	"github.com/leamout/leamout/server/internal/ai"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/integrations/freeswitch"
	natsintegration "github.com/leamout/leamout/server/internal/integrations/nats"
	"github.com/leamout/leamout/server/internal/integrations/postgres"
	redisintegration "github.com/leamout/leamout/server/internal/integrations/redis"
	"github.com/leamout/leamout/server/internal/integrations/ses"
	"github.com/leamout/leamout/server/internal/observability/logging"
	"github.com/leamout/leamout/server/internal/observability/metrics"
	"github.com/leamout/leamout/server/internal/platform/config"
	"github.com/leamout/leamout/server/internal/platform/email"
	"github.com/leamout/leamout/server/internal/platform/idempotency"
	"github.com/leamout/leamout/server/internal/platform/outbox"
	"github.com/leamout/leamout/server/internal/platform/retention"
	platformstorage "github.com/leamout/leamout/server/internal/platform/storage"
	"github.com/leamout/leamout/server/internal/platform/webhooks"
	"github.com/leamout/leamout/server/internal/runtime/agent"
	"github.com/leamout/leamout/server/internal/runtime/calling"
	"github.com/leamout/leamout/server/internal/runtime/medianodes"
	"github.com/leamout/leamout/server/internal/security/encryption"
	"github.com/leamout/leamout/server/internal/telephony/calls"
	"github.com/leamout/leamout/server/internal/telephony/recordings"
	"github.com/leamout/leamout/server/internal/telephony/routing"
	"github.com/leamout/leamout/server/internal/telephony/trunks"
)

type modules struct {
	billingReconciliation       *billing.ReconciliationJob
	billingWebhooks             *billing.WebhookJob
	emailDelivery               *email.Worker
	postgres                    *postgres.Client
	redis                       *redisintegration.Client
	nats                        *natsintegration.Client
	freeSwitch                  *freeswitch.Client
	callsService                *calls.Service
	callConsumer                *calls.Consumer
	agentRuntime                *agent.Runtime
	callReconciliation          *calls.ReconciliationJob
	sessionReconciliation       *sessionReconciliationJob
	toolExecutionReconciliation *toolExecutionReconciliationJob
	sessionRecovery             *sessionRecoveryJob
	outbox                      *outbox.PublisherJob
	webhookConsumer             *webhooks.Consumer
	webhookDelivery             *webhooks.DeliveryJob
	recordingConsumer           *recordings.Consumer
	recordingReconciliation     *recordings.ReconciliationJob
	recordingIngestion          *recordings.IngestionJob
	idempotencyCleanup          *idempotency.CleanupJob
	retentionCleanup            *retention.CleanupJob
	trunkHealth                 *trunks.HealthCheckJob
}

func newModules(ctx context.Context, cfg config.Config, logger *logging.Logger) (*modules, error) {
	postgresClient, err := postgres.New(ctx, postgres.DefaultConfig(cfg.DatabaseURL))
	if err != nil {
		return nil, fmt.Errorf("initialize PostgreSQL: %w", err)
	}

	redisClient, err := redisintegration.New(ctx, redisintegration.DefaultConfig(cfg.RedisURL))
	if err != nil {
		postgresClient.Close()
		return nil, fmt.Errorf("initialize Redis: %w", err)
	}

	natsClient, err := natsintegration.New(ctx, natsintegration.DefaultConfig(cfg.NATSURL))
	if err != nil {
		_ = redisClient.Close()
		postgresClient.Close()
		return nil, fmt.Errorf("initialize NATS: %w", err)
	}
	if err := natsClient.Provision(ctx, natsintegration.DefaultStreamLimits()); err != nil {
		_ = natsClient.Close()
		_ = redisClient.Close()
		postgresClient.Close()
		return nil, fmt.Errorf("provision NATS streams: %w", err)
	}

	freeSwitch, err := freeswitch.New(
		freeswitch.DefaultConfig(cfg.FreeSWITCHESLAddress, cfg.FreeSWITCHESLPassword),
	)
	if err != nil {
		_ = natsClient.Close()
		_ = redisClient.Close()
		postgresClient.Close()
		return nil, fmt.Errorf("initialize FreeSWITCH: %w", err)
	}
	if err := freeSwitch.Connect(ctx); err != nil {
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		_ = redisClient.Close()
		postgresClient.Close()
		return nil, fmt.Errorf("connect FreeSWITCH: %w", err)
	}

	closeDependencies := func() {
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		_ = redisClient.Close()
		postgresClient.Close()
	}

	queries := sqlc.New(postgresClient.Pool())
	credentialCipher, err := encryption.New(cfg.EncryptionKey)
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize Voice Agent tool encryption: %w", err)
	}

	routingRepository := routing.NewRepository(queries, postgresClient.Pool())
	routingService := routing.NewService(routingRepository, nil)
	callsRepository := calls.NewRepository(queries, postgresClient.Pool())
	callController := calling.NewController(freeSwitch)
	admissionLimiter := calling.NewAdmissionLimiter(redisClient)
	callsService := calls.NewService(
		callsRepository,
		routingService,
		callController,
		calling.NewChannelStore(redisClient),
		admissionLimiter,
		metrics.New(redisClient),
	)
	var billingJob *billing.ReconciliationJob
	var billingWebhookJob *billing.WebhookJob
	emailService := email.NewService(credentialCipher)
	cloudCommercial := commercial.New(queries, commercial.Dependencies{
		Database: postgresClient.Pool(),
		Polar:    polar.New(cfg.Polar.AccessToken, cfg.Polar.WebhookSecret, cfg.Polar.Environment),
		PolarProducts: map[string]string{
			"developer": cfg.Polar.DeveloperProductID,
			"pro":       cfg.Polar.ProProductID,
			"scale":     cfg.Polar.ScaleProductID,
		},
		FrontendURL:  cfg.FrontendURL,
		EmailService: emailService,
	})
	planLimits := cloudCommercial.Limits.Service
	if cfg.Polar.AccessToken != "" && cfg.Polar.WebhookSecret != "" {
		billingRepository := billing.NewRepository(queries, postgresClient.Pool())
		billingJob, err = billing.NewReconciliationJob(
			billingRepository,
			cloudCommercial.Billing.Service,
			15*time.Minute,
			func(ctx context.Context, err error) {
				logger.Warn(ctx, "billing reconciliation failed", "error", err)
			},
		)
		if err != nil {
			closeDependencies()
			return nil, err
		}
		billingWebhookJob, err = billing.NewWebhookJob(
			billingRepository,
			cloudCommercial.Billing.Service,
			func(ctx context.Context, err error) {
				logger.Warn(ctx, "billing webhook processing failed", "error", err)
			},
		)
		if err != nil {
			closeDependencies()
			return nil, err
		}
	}

	callsService.SetPlanConcurrencyResolver(planLimits.MaxConcurrentCalls)
	callsService.SetRuntimeAllowanceCheck(planLimits.RequireRuntimeAllowance)
	aiModule := ai.New(queries, ai.Dependencies{
		CredentialCipher: credentialCipher,
		Calls:            callsService,
		Database:         postgresClient.Pool(),
		PlanAgentLimit:   planLimits.MaxVoiceAgents,
	})
	mediaNodes := medianodes.NewRegistry(redisClient)
	aiModule.Orchestration.SetRuntimeAllowanceCheck(planLimits.RequireRuntimeAllowance)
	agentRuntime, err := agent.NewWithMediaNodes(
		aiModule.Orchestration,
		freeSwitch,
		agent.DefaultConfig(cfg.MediaControlURL, cfg.MediaControlToken),
		mediaNodes,
		logger,
	)
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize agent runtime: %w", err)
	}

	callConsumer := calls.NewConsumer(callsService)
	sessionReconciliation := &sessionReconciliationJob{
		queries: queries,
	}
	sessionRecovery := &sessionRecoveryJob{
		queries:    queries,
		freeSwitch: freeSwitch,
		runtime:    agentRuntime,
		logger:     logger,
	}

	callReconciliation, err := calls.NewReconciliationJob(
		callsRepository,
		callsService,
		calls.DefaultReconciliationJobConfig(),
	)
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize call reconciliation: %w", err)
	}

	recordingsRepository := recordings.NewRepository(postgresClient.Pool())
	storageService := platformstorage.NewService(
		platformstorage.NewRepository(queries),
		credentialCipher,
	)
	recordingStorage := platformstorage.NewObjectStorage(storageService)
	recordingsService := recordings.NewService(recordingsRepository, recordingStorage)
	recordingIngestion, err := recordings.NewIngestionJob(
		recordingsRepository,
		recordingStorage,
		recordings.DefaultIngestionConfig(recordings.DefaultStagingPath),
	)
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize recording ingestion: %w", err)
	}
	recordingReconciliation, err := recordings.NewReconciliationJob(
		recordingsRepository,
		recordings.DefaultReconciliationJobConfig(),
	)
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize recording reconciliation: %w", err)
	}
	retentionCleanup, err := retention.NewCleanupJob(
		retention.NewRepository(queries),
		recordingsService,
		retention.DefaultCleanupJobConfig(),
	)
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize retention cleanup: %w", err)
	}

	idempotencyCleanup, err := idempotency.NewCleanupJob(
		idempotency.NewRepository(queries),
		time.Hour,
	)
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize idempotency cleanup: %w", err)
	}

	policy, err := cfg.SIPEgressPolicy()
	if err != nil {
		closeDependencies()
		return nil, err
	}
	callsService.SetEgressPolicy(policy)
	trunkHealth, err := trunks.NewHealthCheckJob(
		trunks.NewRepository(queries),
		trunks.SIPOptionsProber{Policy: policy},
		trunks.DefaultHealthCheckConfig(),
	)
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize trunk health checks: %w", err)
	}

	outboxJob, err := outbox.NewPublisherJob(
		outbox.NewRepository(queries),
		outbox.NewPublisher(natsClient),
		outbox.DefaultPublisherJobConfig("worker-outbox"),
	)
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize outbox publisher: %w", err)
	}

	webhookRepository := webhooks.NewRepository(queries)
	webhookService := webhooks.NewService(webhookRepository, postgresClient.Pool())
	webhookDeliveryJob, err := webhooks.NewDeliveryJob(
		webhookRepository,
		webhooks.NewHTTPSender(),
		webhooks.DefaultDeliveryJobConfig("worker-webhooks"),
	)
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize webhook delivery worker: %w", err)
	}

	sender, err := ses.New(ctx, ses.Config{
		Region:           cfg.AWS.Region,
		From:             cfg.AWS.FromEmail,
		AccessKey:        cfg.AWS.AccessKey,
		SecretKey:        cfg.AWS.SecretKey,
		ConfigurationSet: cfg.AWS.ConfigurationSet,
	})
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize SES: %w", err)
	}
	emailDelivery := email.NewWorker(email.NewRepository(postgresClient.Pool()), sender, credentialCipher)
	return &modules{
		billingReconciliation:       billingJob,
		billingWebhooks:             billingWebhookJob,
		emailDelivery:               emailDelivery,
		postgres:                    postgresClient,
		redis:                       redisClient,
		nats:                        natsClient,
		freeSwitch:                  freeSwitch,
		callsService:                callsService,
		callConsumer:                callConsumer,
		agentRuntime:                agentRuntime,
		callReconciliation:          callReconciliation,
		sessionReconciliation:       sessionReconciliation,
		sessionRecovery:             sessionRecovery,
		toolExecutionReconciliation: &toolExecutionReconciliationJob{queries: queries},
		outbox:                      outboxJob,
		webhookConsumer:             webhooks.NewConsumer(natsClient, webhookService),
		webhookDelivery:             webhookDeliveryJob,
		recordingConsumer:           recordings.NewConsumer(recordingsService),
		recordingReconciliation:     recordingReconciliation,
		recordingIngestion:          recordingIngestion,
		idempotencyCleanup:          idempotencyCleanup,
		retentionCleanup:            retentionCleanup,
		trunkHealth:                 trunkHealth,
	}, nil
}

func (m *modules) close(logger *logging.Logger) {
	if m == nil {
		return
	}
	if m.freeSwitch != nil {
		if err := m.freeSwitch.Close(); err != nil {
			logger.Warn(context.Background(), "close FreeSWITCH", "error", err)
		}
	}
	if m.nats != nil {
		if err := m.nats.Close(); err != nil {
			logger.Warn(context.Background(), "close NATS", "error", err)
		}
	}
	if m.redis != nil {
		if err := m.redis.Close(); err != nil {
			logger.Warn(context.Background(), "close Redis", "error", err)
		}
	}
	if m.postgres != nil {
		m.postgres.Close()
	}
}
