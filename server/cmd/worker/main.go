package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/leamout/leamout/server/internal/platform/logging"
	runtimeworker "github.com/leamout/leamout/server/internal/runtime/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := runtimeworker.Run(ctx); err != nil {
		logging.New().Error(context.Background(), "worker failed", "error", err)
		os.Exit(1)
	}
}
