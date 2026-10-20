package backoffice

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/leamout/leamout/server/internal/observability/logging"
	"github.com/leamout/leamout/server/internal/platform/config"
)

func Run(ctx context.Context) error {
	logger := logging.New().With("process", "backoffice")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	if err := cfg.ValidateForBackoffice(); err != nil {
		return fmt.Errorf("validate configuration: %w", err)
	}

	modules, err := newModules(ctx, cfg, logger)
	if err != nil {
		return err
	}
	defer modules.postgres.Close()

	router, err := newRouter(modules, logger, cfg.BackofficeURL)
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:              ":8082",
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info(ctx, "backoffice listening", "address", httpServer.Addr)
		err := httpServer.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		serverErr <- err
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
	}

	logger.Info(context.Background(), "backoffice shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}
	if err := <-serverErr; err != nil {
		return fmt.Errorf("serve HTTP: %w", err)
	}
	return nil
}
