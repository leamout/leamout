package worker

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/commercial"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/integrations/carriers/commpeak"
	"github.com/leamout/leamout/internal/integrations/carriers/didww"
	"github.com/leamout/leamout/internal/integrations/freeswitch"
	natsintegration "github.com/leamout/leamout/internal/integrations/nats"
	redisintegration "github.com/leamout/leamout/internal/integrations/redis"
	"github.com/leamout/leamout/internal/modules/idempotency"
	"github.com/leamout/leamout/internal/modules/outbox"
	"github.com/leamout/leamout/internal/modules/webhooks"
	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/telecom/numbers"
	"github.com/leamout/leamout/internal/telecom/wholesale"
)

type dependencies struct {
	db         *pgxpool.Pool
	freeSwitch *freeswitch.Client
	nats       *natsintegration.Client
	redis      *redisintegration.Client
	queries    *sqlc.Queries
}

func newDependencies(ctx context.Context, cfg config.CloudConfig) (*dependencies, error) {
	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, wrapWorkerError("connect worker database", err)
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, wrapWorkerError("ping worker database", err)
	}

	natsClient, err := natsintegration.New(ctx, natsintegration.DefaultConfig(cfg.NATSURL))
	if err != nil {
		db.Close()
		return nil, wrapWorkerError("connect worker NATS", err)
	}
	redisClient, err := redisintegration.New(ctx, redisintegration.DefaultConfig(cfg.RedisURL))
	if err != nil {
		_ = natsClient.Close()
		db.Close()
		return nil, wrapWorkerError("connect worker Redis", err)
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
		return nil, wrapWorkerError("provision worker NATS streams", err)
	}

	freeSwitch, err := freeswitch.New(freeswitch.DefaultConfig(cfg.FreeSWITCHESLAddress, cfg.FreeSWITCHESLPassword))
	if err != nil {
		_ = redisClient.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, wrapWorkerError("initialize worker FreeSWITCH client", err)
	}
	if err := freeSwitch.Connect(ctx); err != nil {
		_ = redisClient.Close()
		_ = freeSwitch.Close()
		_ = natsClient.Close()
		db.Close()
		return nil, wrapWorkerError("connect worker FreeSWITCH", err)
	}

	return &dependencies{
		db:         db,
		freeSwitch: freeSwitch,
		nats:       natsClient,
		redis:      redisClient,
		queries:    sqlc.New(db),
	}, nil
}

func (d *dependencies) close() {
	if d.freeSwitch != nil {
		_ = d.freeSwitch.Close()
	}
	if d.nats != nil {
		_ = d.nats.Close()
	}
	if d.redis != nil {
		_ = d.redis.Close()
	}
	if d.db != nil {
		d.db.Close()
	}
}

type sharedJobs struct {
	calls      *callJobs
	recordings *recordingJobs
}

func newSharedJobs(deps *dependencies) (*sharedJobs, error) {
	calls, err := newCallJobs(deps)
	if err != nil {
		return nil, err
	}
	recordings, err := newRecordingJobs(deps)
	if err != nil {
		return nil, err
	}
	return &sharedJobs{calls: calls, recordings: recordings}, nil
}

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

func newIdempotencyCleanup(deps *dependencies) (*idempotency.CleanupJob, error) {
	job, err := idempotency.NewCleanupJob(
		idempotency.NewRepository(deps.queries),
		time.Hour,
	)
	if err != nil {
		return nil, wrapWorkerError("initialize idempotency cleanup job", err)
	}
	return job, nil
}

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

func newManagedJobs(
	db *pgxpool.Pool,
	redisClient *redisintegration.Client,
	cfg config.CloudConfig,
) (*numbers.ProviderOperationJob, *wholesale.CDRPollJob, error) {
	commercialModule := commercial.NewCloud(db)
	numbersRepository := numbers.NewRepository(db, redisClient)
	numbersService := numbers.NewService(numbersRepository)
	numbersService.SetManagedPurchaseAuthority(commercialModule.Wallets.Service)

	if strings.TrimSpace(cfg.DIDWW.APIKey) != "" {
		didwwClient, err := didww.NewClient(didww.Config{
			BaseURL: cfg.DIDWW.APIBaseURL,
			APIKey:  cfg.DIDWW.APIKey,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("initialize DIDWW provider executor: %w", err)
		}
		numbersService.SetManagedProvider("didww", didwwClient)
	}
	providerOperations, err := numbers.NewProviderOperationJob(
		numbersRepository,
		numbersService,
		numbers.DefaultProviderOperationJobConfig(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize provider operation job: %w", err)
	}

	var commpeakSource wholesale.CDRPageSource
	if strings.TrimSpace(cfg.CommPeak.Authorization) != "" {
		commpeakClient, err := commpeak.NewClient(commpeak.Config{
			BaseURL:       cfg.CommPeak.APIBaseURL,
			Authorization: cfg.CommPeak.Authorization,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("initialize CommPeak CDR client: %w", err)
		}
		commpeakSource = commpeakClient
	}
	commpeakCDRPolling, err := wholesale.NewCDRPollJob(
		wholesale.NewRepository(db),
		commpeakSource,
		wholesale.DefaultCDRPollJobConfig("commpeak"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("initialize CommPeak CDR polling job: %w", err)
	}
	return providerOperations, commpeakCDRPolling, nil
}
