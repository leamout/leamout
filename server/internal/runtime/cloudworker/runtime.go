package worker

import (
	"context"

	"github.com/leamout/leamout/internal/platform/config"
)

func New(ctx context.Context, cfg config.Config) (*Worker, error) {
	return NewCloud(ctx, cfg)
}
