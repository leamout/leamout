package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/leamout/leamout/internal/platform/config"
	selfhostedworkerruntime "github.com/leamout/leamout/internal/runtime/selfhostedworker"
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

	worker, err := selfhostedworkerruntime.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer worker.Close()

	log.Print("self-hosted worker started")
	if err := worker.Run(ctx); err != nil {
		log.Fatal(err)
	}
	log.Print("self-hosted worker stopped")
}
