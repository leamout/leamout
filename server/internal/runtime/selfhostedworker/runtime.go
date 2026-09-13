package selfhostedworker

import (
	"context"

	appworker "github.com/leamout/leamout/internal/app/worker"
	"github.com/leamout/leamout/internal/platform/config"
)

func New(ctx context.Context, cfg config.Config) (*appworker.Worker, error) {
	return appworker.NewSelfHosted(ctx, cfg)
}
