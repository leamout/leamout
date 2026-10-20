package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"

	"github.com/leamout/leamout/server/internal/integrations/keycloak"
)

// BackofficeConfig loads only dependencies used by the operator HTTP process.
type BackofficeConfig struct {
	Keycloak      keycloak.Config `envPrefix:"KEYCLOAK_"`
	AppEnv        string          `env:"APP_ENV" envDefault:"development"`
	FrontendURL   string          `env:"FRONTEND_URL" envDefault:"http://localhost:3000"`
	DatabaseURL   string          `env:"DATABASE_URL,required"`
	EncryptionKey string          `env:"ENCRYPTION_KEY,required"`
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
	cfg.EncryptionKey = strings.TrimSpace(cfg.EncryptionKey)
	cfg.FrontendURL = strings.TrimRight(strings.TrimSpace(cfg.FrontendURL), "/")
	if cfg.DatabaseURL == "" || cfg.EncryptionKey == "" {
		return BackofficeConfig{}, fmt.Errorf("database URL and encryption key are required")
	}
	return cfg, nil
}

func (c BackofficeConfig) IsDevelopment() bool {
	return strings.EqualFold(strings.TrimSpace(c.AppEnv), "development")
}
