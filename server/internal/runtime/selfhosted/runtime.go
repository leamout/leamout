package selfhosted

import (
	"context"

	appserver "github.com/leamout/leamout/internal/app/server"
	"github.com/leamout/leamout/internal/platform/config"
)

func New(ctx context.Context, cfg config.Config) (*appserver.Server, error) {
	return appserver.NewSelfHosted(ctx, cfg)
}
