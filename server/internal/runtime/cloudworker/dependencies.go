package worker

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/integrations/freeswitch"
	natsintegration "github.com/leamout/leamout/internal/integrations/nats"
	redisintegration "github.com/leamout/leamout/internal/integrations/redis"
	"github.com/leamout/leamout/internal/platform/config"
)

type dependencies struct {
	db         *pgxpool.Pool
	freeSwitch *freeswitch.Client
	nats       *natsintegration.Client
	redis      *redisintegration.Client
	queries    *sqlc.Queries
}

func newDependencies(ctx context.Context, cfg config.Config) (*dependencies, error) {
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
