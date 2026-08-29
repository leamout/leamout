package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/integrations/freeswitch"
	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/routing"
)

type Worker struct {
	db         *pgxpool.Pool
	freeSwitch *freeswitch.Client
	calls      *calls.Consumer
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

	freeSwitch, err := freeswitch.New(
		freeswitch.DefaultConfig(cfg.FreeSWITCHESLAddress, cfg.FreeSWITCHESLPassword),
	)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize worker FreeSWITCH client: %w", err)
	}
	if err := freeSwitch.Connect(ctx); err != nil {
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

	return &Worker{
		db:         db,
		freeSwitch: freeSwitch,
		calls:      calls.NewConsumer(callsService),
	}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	events := []string{
		"CHANNEL_CREATE",
		"CHANNEL_ANSWER",
		"CHANNEL_HANGUP_COMPLETE",
	}

	if err := w.freeSwitch.Subscribe(ctx, freeswitch.EventFormatPlain, events, func(eventCtx context.Context, event freeswitch.Event) error {
		if err := w.calls.HandleFreeSWITCHEvent(eventCtx, event); err != nil {
			log.Printf("FreeSWITCH call event %s failed: %v", event.Name, err)
			return err
		}
		return nil
	}); err != nil {
		return fmt.Errorf("subscribe to FreeSWITCH call events: %w", err)
	}

	log.Print("worker subscribed to FreeSWITCH call lifecycle events")
	<-ctx.Done()
	return nil
}

func (w *Worker) Close() {
	if w.freeSwitch != nil {
		_ = w.freeSwitch.Close()
	}
	if w.db != nil {
		w.db.Close()
	}
}
