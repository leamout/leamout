package server

import (
	"context"

	"github.com/leamout/leamout/internal/platform/config"
)

func New(ctx context.Context, cfg config.Config) (*Server, error) {
	return NewSelfHosted(ctx, cfg)
}
