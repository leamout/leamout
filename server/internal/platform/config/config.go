package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv                     string        `env:"APP_ENV" envDefault:"development"`
	DatabaseURL                string        `env:"DATABASE_URL,required"`
	RedisURL                   string        `env:"REDIS_URL,required"`
	NATSURL                    string        `env:"NATS_URL,required"`
	NATSStreamReplicas         int           `env:"NATS_STREAM_REPLICAS" envDefault:"1"`
	FreeSWITCHESLAddress       string        `env:"FREESWITCH_ESL_ADDRESS" envDefault:"127.0.0.1:8021"`
	FreeSWITCHESLPassword      string        `env:"FREESWITCH_ESL_PASSWORD,required"`
	CarrierCredentialKey       string        `env:"CARRIER_CREDENTIAL_ENCRYPTION_KEY,required"`
	TURNAuthSecret             string        `env:"TURN_AUTH_SECRET,required"`
	TURNPublicURLs             []string      `env:"TURN_PUBLIC_URLS" envSeparator:"," envDefault:"stun:localhost:3478,turn:localhost:3478?transport=udp,turn:localhost:3478?transport=tcp"`
	TURNCredentialTTL          time.Duration `env:"TURN_CREDENTIAL_TTL" envDefault:"10m"`
	TURNCredentialIssueLimit   int64         `env:"TURN_CREDENTIAL_ISSUE_LIMIT" envDefault:"60"`
	TURNCredentialIssueWindow  time.Duration `env:"TURN_CREDENTIAL_ISSUE_WINDOW" envDefault:"1m"`
	EndpointHealthInterval     time.Duration `env:"ENDPOINT_HEALTH_INTERVAL" envDefault:"10s"`
	EndpointHealthProbeTimeout time.Duration `env:"ENDPOINT_HEALTH_PROBE_TIMEOUT" envDefault:"2s"`
	EndpointHealthCooldown     time.Duration `env:"ENDPOINT_HEALTH_COOLDOWN" envDefault:"30s"`
	EndpointHealthFailures     int32         `env:"ENDPOINT_HEALTH_FAILURE_THRESHOLD" envDefault:"3"`
	EndpointHealthBatchSize    int32         `env:"ENDPOINT_HEALTH_BATCH_SIZE" envDefault:"100"`
	EndpointHealthConcurrency  int           `env:"ENDPOINT_HEALTH_CONCURRENCY" envDefault:"10"`
	CORSOrigins                []string      `env:"CORS_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000,http://127.0.0.1:3000"`
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
