package config

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type MinIOConfig struct {
	AccessKey string `env:"APP_ACCESS_KEY,required"`
	SecretKey string `env:"APP_SECRET_KEY,required"`
}

type StripeConfig struct {
	SecretKey        string `env:"SECRET_KEY"`
	WebhookSecret    string `env:"WEBHOOK_SECRET"`
	DeveloperPriceID string `env:"DEVELOPER_PRICE_ID"`
	ProPriceID       string `env:"PRO_PRICE_ID"`
}

type AWSConfig struct {
	FromEmail        string `env:"FROM_EMAIL" envDefault:"noreply@leamout.com"`
	Region           string `env:"REGION" envDefault:"us-east-1"`
	AccessKey        string `env:"ACCESS_KEY_ID"`
	SecretKey        string `env:"SECRET_ACCESS_KEY"`
	ConfigurationSet string `env:"CONFIGURATION_SET" envDefault:"leamout-transactional"`
}

type Config struct {
	AWS    AWSConfig    `envPrefix:"AWS_"`
	Stripe StripeConfig `envPrefix:"STRIPE_"`

	AppEnv                string      `env:"APP_ENV" envDefault:"development"`
	Domain                string      `env:"DOMAIN"`
	DatabaseURL           string      `env:"DATABASE_URL,required"`
	RedisURL              string      `env:"REDIS_URL,required"`
	NATSURL               string      `env:"NATS_URL,required"`
	FreeSWITCHESLAddress  string      `env:"FREESWITCH_ESL_ADDRESS" envDefault:"127.0.0.1:8021"`
	FreeSWITCHESLPassword string      `env:"FREESWITCH_ESL_PASSWORD,required"`
	MediaControlURL       string      `env:"MEDIA_CONTROL_URL" envDefault:"http://127.0.0.1:8090"`
	MediaControlToken     string      `env:"MEDIA_CONTROL_TOKEN"`
	OpenAIAPIKey          string      `env:"OPENAI_API_KEY"`
	GeminiAPIKey          string      `env:"GEMINI_API_KEY"`
	DeepgramAPIKey        string      `env:"DEEPGRAM_API_KEY"`
	AssemblyAIAPIKey      string      `env:"ASSEMBLYAI_API_KEY"`
	GroqAPIKey            string      `env:"GROQ_API_KEY"`
	CartesiaAPIKey        string      `env:"CARTESIA_API_KEY"`
	ElevenLabsAPIKey      string      `env:"ELEVENLABS_API_KEY"`
	EncryptionKey         string      `env:"ENCRYPTION_KEY,required"`
	MinIO                 MinIOConfig `envPrefix:"MINIO_"`
	TURNAuthSecret        string      `env:"TURN_AUTH_SECRET,required"`
	TURNPublicURLs        []string    `env:"TURN_PUBLIC_URLS" envSeparator:"," envDefault:"stun:localhost:3478,turn:localhost:3478?transport=udp,turn:localhost:3478?transport=tcp"`
	CORSOrigins           []string    `env:"CORS_ORIGINS" envSeparator:"," envDefault:"http://localhost:3000,http://127.0.0.1:3000"`
	TrustedProxyCIDRs     []string    `env:"TRUSTED_PROXY_CIDRS" envSeparator:","`
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

	return cfg, nil
}

func (c Config) IsDevelopment() bool {
	return strings.EqualFold(c.AppEnv, "development")
}

func (c *Config) normalize() {
	c.AWS.Region = strings.TrimSpace(c.AWS.Region)
	c.AWS.FromEmail = strings.TrimSpace(c.AWS.FromEmail)
	c.AWS.ConfigurationSet = strings.TrimSpace(c.AWS.ConfigurationSet)
	c.Stripe.SecretKey = strings.TrimSpace(c.Stripe.SecretKey)
	c.Stripe.WebhookSecret = strings.TrimSpace(c.Stripe.WebhookSecret)
	c.Stripe.DeveloperPriceID = strings.TrimSpace(c.Stripe.DeveloperPriceID)
	c.Stripe.ProPriceID = strings.TrimSpace(c.Stripe.ProPriceID)
	c.AppEnv = strings.TrimSpace(c.AppEnv)
	c.Domain = strings.TrimSpace(c.Domain)
	c.DatabaseURL = strings.TrimSpace(c.DatabaseURL)
	c.RedisURL = strings.TrimSpace(c.RedisURL)
	c.NATSURL = strings.TrimSpace(c.NATSURL)
	c.FreeSWITCHESLAddress = strings.TrimSpace(c.FreeSWITCHESLAddress)
	c.FreeSWITCHESLPassword = strings.TrimSpace(c.FreeSWITCHESLPassword)
	c.MediaControlURL = strings.TrimRight(strings.TrimSpace(c.MediaControlURL), "/")
	c.MediaControlToken = strings.TrimSpace(c.MediaControlToken)
	c.OpenAIAPIKey = strings.TrimSpace(c.OpenAIAPIKey)
	c.GeminiAPIKey = strings.TrimSpace(c.GeminiAPIKey)
	c.DeepgramAPIKey = strings.TrimSpace(c.DeepgramAPIKey)
	c.AssemblyAIAPIKey = strings.TrimSpace(c.AssemblyAIAPIKey)
	c.GroqAPIKey = strings.TrimSpace(c.GroqAPIKey)
	c.CartesiaAPIKey = strings.TrimSpace(c.CartesiaAPIKey)
	c.ElevenLabsAPIKey = strings.TrimSpace(c.ElevenLabsAPIKey)
	c.EncryptionKey = strings.TrimSpace(c.EncryptionKey)
	c.MinIO.AccessKey = strings.TrimSpace(c.MinIO.AccessKey)
	c.MinIO.SecretKey = strings.TrimSpace(c.MinIO.SecretKey)
	c.TURNAuthSecret = strings.TrimSpace(c.TURNAuthSecret)
	c.TURNPublicURLs = normalizeStrings(c.TURNPublicURLs)
	c.CORSOrigins = normalizeStrings(c.CORSOrigins)
	c.TrustedProxyCIDRs = normalizeStrings(c.TrustedProxyCIDRs)
}

func (c Config) TrustedProxyPrefixes() ([]netip.Prefix, error) {
	result := make([]netip.Prefix, 0, len(c.TrustedProxyCIDRs))
	for _, value := range c.TrustedProxyCIDRs {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return nil, fmt.Errorf("parse trusted proxy CIDR %q: %w", value, err)
		}
		result = append(result, prefix.Masked())
	}
	return result, nil
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
