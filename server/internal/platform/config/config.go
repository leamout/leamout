package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                string   `env:"APP_ENV" envDefault:"development"`
	DatabaseURL           string   `env:"DATABASE_URL,required"`
	RedisURL              string   `env:"REDIS_URL,required"`
	NATSURL               string   `env:"NATS_URL,required"`
	NATSStreamReplicas    int      `env:"NATS_STREAM_REPLICAS" envDefault:"1"`
	FreeSWITCHESLAddress  string   `env:"FREESWITCH_ESL_ADDRESS" envDefault:"127.0.0.1:8021"`
	FreeSWITCHESLPassword string   `env:"FREESWITCH_ESL_PASSWORD,required"`
	CarrierCredentialKey  string   `env:"CARRIER_CREDENTIAL_ENCRYPTION_KEY,required"`
	TURNAddress           string   `env:"TURN_ADDRESS" envDefault:"127.0.0.1:3478"`
	TURNAuthSecret        string   `env:"TURN_AUTH_SECRET,required"`
	TURNPublicURLs        []string `env:"TURN_PUBLIC_URLS" envSeparator:"," envDefault:"stun:localhost:3478,turn:localhost:3478?transport=udp,turn:localhost:3478?transport=tcp"`
	CORSOrigins           []string `env:"CORS_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000,http://127.0.0.1:3000"`
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse environment: %w", err)
	}

	return cfg, nil
}

func (c Config) IsDevelopment() bool {
	return strings.EqualFold(c.AppEnv, "development")
}
