package main

import (
	"context"
	"github.com/leamout/leamout/server/internal/platform/logging"
	exporter "github.com/leamout/leamout/server/internal/runtime/freeswitchexporter"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := exporter.Run(ctx); err != nil {
		logging.New().Error(context.Background(), "freeswitch exporter failed", "error", err)
		os.Exit(1)
	}
}
