package selfhosted

import (
	"context"

	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/runtime/server"
	"github.com/leamout/leamout/internal/runtime/worker"
)

func NewServer(ctx context.Context, cfg config.Config) (*server.Server, error) {
	return server.NewSelfHosted(ctx, cfg)
}

func NewWorker(ctx context.Context, cfg config.Config) (*worker.Worker, error) {
	return worker.NewSelfHosted(ctx, cfg)
}
