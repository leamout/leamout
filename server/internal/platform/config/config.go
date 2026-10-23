package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"github.com/leamout/leamout/server/internal/integrations/kinde"
)

type PolarConfig struct {
	Environment        string `env:"ENVIRONMENT" envDefault:"production"`
	AccessToken        string `env:"ACCESS_TOKEN"`
	WebhookSecret      string `env:"WEBHOOK_SECRET"`
	DeveloperProductID string `env:"DEVELOPER_PRODUCT_ID" envDefault:"58ff657a-74d9-4342-89c4-2a3b784c6965"`
	ProProductID       string `env:"PRO_PRODUCT_ID" envDefault:"da5b3818-e92e-4b09-afad-ae61895f05db"`
	ScaleProductID     string `env:"SCALE_PRODUCT_ID" envDefault:"a3023020-9822-4835-9509-1635c7d35f63"`
}

type AWSConfig struct {
	FromEmail        string `env:"FROM_EMAIL" envDefault:"noreply@leamout.com"`
	Region           string `env:"REGION" envDefault:"us-east-1"`
	AccessKey        string `env:"ACCESS_KEY_ID"`
	SecretKey        string `env:"SECRET_ACCESS_KEY"`
	ConfigurationSet string `env:"CONFIGURATION_SET" envDefault:"leamout-transactional"`
}

type Config struct {
	Kinde    kinde.Config `envPrefix:"KINDE_"`
	AWS      AWSConfig       `envPrefix:"AWS_"`
	Polar    PolarConfig     `envPrefix:"POLAR_"`

	AppEnv                string   `env:"APP_ENV" envDefault:"development"`
	BackofficeURL         string   `env:"BACKOFFICE_URL" envDefault:"http://localhost:8082"`
	FrontendURL           string   `env:"FRONTEND_URL" envDefault:"http://localhost:3000"`
	BackendURL            string   `env:"BACKEND_URL" envDefault:"http://localhost:8080"`
	Domain                string   `env:"DOMAIN"`
	DatabaseURL           string   `env:"DATABASE_URL"`
	RedisURL              string   `env:"REDIS_URL"`
	NATSURL               string   `env:"NATS_URL"`
	FreeSWITCHESLAddress  string   `env:"FREESWITCH_ESL_ADDRESS" envDefault:"127.0.0.1:8021"`
	FreeSWITCHESLPassword string   `env:"FREESWITCH_ESL_PASSWORD"`
	MediaControlURL       string   `env:"MEDIA_CONTROL_URL" envDefault:"http://127.0.0.1:8090"`
	MediaControlToken     string   `env:"MEDIA_CONTROL_TOKEN"`
	EncryptionKey         string   `env:"ENCRYPTION_KEY"`
	TURNAuthSecret        string   `env:"TURN_AUTH_SECRET"`
	TURNPublicURLs        []string `env:"TURN_PUBLIC_URLS" envSeparator:"," envDefault:"stun:localhost:3478,turn:localhost:3478?transport=udp,turn:localhost:3478?transport=tcp"`
	CORSOrigins           []string `env:"CORS_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000,http://127.0.0.1:3000"`
	TrustedProxyCIDRs     []string `env:"TRUSTED_PROXY_CIDRS" envSeparator:","`
	SIPAllowedEgressIPs   []string `env:"SIP_ALLOWED_EGRESS_IPS" envSeparator:","`
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}

	cfg, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, fmt.Errorf("parse environment: %w", err)
	}

	cfg.normalize()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) IsDevelopment() bool {
	return strings.EqualFold(c.AppEnv, "development")
}

func (c *Config) normalize() {
	c.AWS.Region = strings.TrimSpace(c.AWS.Region)
	c.AWS.FromEmail = strings.TrimSpace(c.AWS.FromEmail)
	c.AWS.ConfigurationSet = strings.TrimSpace(c.AWS.ConfigurationSet)
	c.Polar.Environment = strings.TrimSpace(c.Polar.Environment)
	c.Polar.AccessToken = strings.TrimSpace(c.Polar.AccessToken)
	c.Polar.WebhookSecret = strings.TrimSpace(c.Polar.WebhookSecret)
	c.Polar.DeveloperProductID = strings.TrimSpace(c.Polar.DeveloperProductID)
	c.Polar.ProProductID = strings.TrimSpace(c.Polar.ProProductID)
	c.Polar.ScaleProductID = strings.TrimSpace(c.Polar.ScaleProductID)
	c.AppEnv = strings.TrimSpace(c.AppEnv)
	c.Domain = strings.TrimSpace(c.Domain)
	c.BackofficeURL = strings.TrimRight(strings.TrimSpace(c.BackofficeURL), "/")
	c.FrontendURL = strings.TrimRight(strings.TrimSpace(c.FrontendURL), "/")
	c.BackendURL = strings.TrimRight(strings.TrimSpace(c.BackendURL), "/")
	c.DatabaseURL = strings.TrimSpace(c.DatabaseURL)
	c.RedisURL = strings.TrimSpace(c.RedisURL)
	c.NATSURL = strings.TrimSpace(c.NATSURL)
	c.FreeSWITCHESLAddress = strings.TrimSpace(c.FreeSWITCHESLAddress)
	c.FreeSWITCHESLPassword = strings.TrimSpace(c.FreeSWITCHESLPassword)
	c.MediaControlURL = strings.TrimRight(strings.TrimSpace(c.MediaControlURL), "/")
	c.MediaControlToken = strings.TrimSpace(c.MediaControlToken)
	c.EncryptionKey = strings.TrimSpace(c.EncryptionKey)
	c.TURNAuthSecret = strings.TrimSpace(c.TURNAuthSecret)
	c.TURNPublicURLs = normalizeStrings(c.TURNPublicURLs)
	c.CORSOrigins = normalizeStrings(c.CORSOrigins)
	c.TrustedProxyCIDRs = normalizeStrings(c.TrustedProxyCIDRs)
	c.SIPAllowedEgressIPs = normalizeStrings(c.SIPAllowedEgressIPs)
}

func normalizeStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			result = append(result, value)
		}
	}
	return result
}
