package worker

import (
	"context"
	"fmt"

	"github.com/leamout/leamout/server/internal/platform/config"
	"github.com/leamout/leamout/server/internal/platform/logging"
)

// Run starts the license-free Core worker.
func Run(ctx context.Context) error {
	return run(ctx, false)
}

// RunCloud starts the worker with Cloud commercial policies installed.
func RunCloud(ctx context.Context) error {
	return run(ctx, true)
}

func run(ctx context.Context, cloud bool) error {
	logger := logging.New().With("process", "worker")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	if cloud {
		if err := cfg.Polar.Validate(); err != nil {
			return fmt.Errorf("validate Cloud billing config: %w", err)
		}
	}
	modules, err := newModules(ctx, cfg, logger, cloud)
	if err != nil {
		return err
	}
	defer modules.close(logger)

	if err := modules.ready(ctx); err != nil {
		return fmt.Errorf("worker readiness: %w", err)
	}
	if err := subscribeFreeSWITCH(ctx, logger, modules); err != nil {
		return fmt.Errorf("subscribe FreeSWITCH events: %w", err)
	}

	if err := runWorkloadsWithReadiness(ctx, logger, modules); err != nil {
		return fmt.Errorf("worker runtime: %w", err)
	}

	logger.Info(context.Background(), "worker stopped")
	return nil
}
