package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/leamout/leamout/internal/platform/config"
	runtimeworker "github.com/leamout/leamout/internal/runtime/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	worker, err := runtimeworker.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer worker.Close()

	log.Print("worker started")
	if err := worker.Run(ctx); err != nil {
		log.Fatal(err)
	}
	log.Print("worker stopped")
}
