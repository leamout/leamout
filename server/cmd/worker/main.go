package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/leamout/leamout/internal/platform/config"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if _, err := config.Load(); err != nil {
		log.Fatal(err)
	}

	log.Print("worker started")
	<-ctx.Done()
	log.Print("worker stopped")
}
