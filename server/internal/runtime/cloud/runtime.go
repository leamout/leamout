package cloud

import (
	"context"

	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/runtime/server"
	"github.com/leamout/leamout/internal/runtime/worker"
)

func NewServer(ctx context.Context, cfg config.Config) (*server.Server, error) {
	return server.NewCloud(ctx, cfg)
}

func NewWorker(ctx context.Context, cfg config.Config) (*worker.Worker, error) {
	return worker.NewCloud(ctx, cfg)
}
