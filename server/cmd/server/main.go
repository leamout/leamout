package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/runtime/server"
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

	srv, err := server.New(ctx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Close()

	httpServer := &http.Server{
		Addr:              ":8080",
		Handler:           srv.Router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	serverErr := make(chan error, 1)

	go func() {
		log.Printf("server listening on %s", httpServer.Addr)
		serverErr <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}

	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}
}
