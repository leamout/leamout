package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// BackofficeConfig contains only the process configuration required by the
// operator-facing Backoffice runtime.
type BackofficeConfig struct {
	DatabaseURL string `env:"DATABASE_URL,required"`
}

func LoadBackoffice() (BackofficeConfig, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return BackofficeConfig{}, fmt.Errorf("load .env: %w", err)
	}

	cfg, err := env.ParseAs[BackofficeConfig]()
	if err != nil {
		return BackofficeConfig{}, fmt.Errorf("parse backoffice environment: %w", err)
	}
	cfg.DatabaseURL = strings.TrimSpace(cfg.DatabaseURL)
	return cfg, nil
}
