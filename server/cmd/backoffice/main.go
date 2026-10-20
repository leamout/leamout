package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/leamout/leamout/server/internal/observability/logging"
	runtimebackoffice "github.com/leamout/leamout/server/internal/runtime/backoffice"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runtimebackoffice.Run(ctx); err != nil {
		logging.New().Error(context.Background(), "backoffice failed", "error", err)
		os.Exit(1)
	}
}
