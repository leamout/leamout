package cloudworker

import (
	"context"

	appworker "github.com/leamout/leamout/internal/app/worker"
	"github.com/leamout/leamout/internal/platform/config"
)

func New(ctx context.Context, cfg config.Config) (*appworker.Worker, error) {
	return appworker.NewCloud(ctx, cfg)
}
