package server

import (
	"context"
	"fmt"

	"github.com/leamout/leamout/server/internal/ai"
	"github.com/leamout/leamout/server/internal/ai/providers"
	"github.com/leamout/leamout/server/internal/commercial"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/identity"
	"github.com/leamout/leamout/server/internal/integrations/coturn"
	"github.com/leamout/leamout/server/internal/integrations/freeswitch"
	"github.com/leamout/leamout/server/internal/integrations/minio"
	"github.com/leamout/leamout/server/internal/integrations/payments/stripe"
	"github.com/leamout/leamout/server/internal/integrations/postgres"
	redisintegration "github.com/leamout/leamout/server/internal/integrations/redis"
	"github.com/leamout/leamout/server/internal/platform"
	"github.com/leamout/leamout/server/internal/platform/config"
	"github.com/leamout/leamout/server/internal/platform/logging"
	"github.com/leamout/leamout/server/internal/platform/metrics"
	"github.com/leamout/leamout/server/internal/platform/middleware"
	"github.com/leamout/leamout/server/internal/runtime/calling"
	"github.com/leamout/leamout/server/internal/security/authn"
	"github.com/leamout/leamout/server/internal/security/encryption"
	"github.com/leamout/leamout/server/internal/telephony"
	"github.com/leamout/leamout/server/internal/telephony/recordings"
	"github.com/leamout/leamout/server/internal/telephony/webrtc"
	"github.com/leamout/leamout/server/internal/tenancy"
)

type modules struct {
	postgres             *postgres.Client
	redis                *redisintegration.Client
	freeSwitch           *freeswitch.Client
	identity             *identity.Module
	tenancy              *tenancy.Module
	platform             *platform.Module
	commercial           *commercial.Module
	ai                   *ai.Module
	telephony            *telephony.Module
	authn                *middleware.AuthnMiddleware
	organizationsContext *middleware.OrganizationMiddleware
	rateLimit            *middleware.RateLimitMiddleware
	metrics              *metrics.Registry
	queries              *sqlc.Queries
	credentialCipher     *encryption.Cipher
}

func newModules(ctx context.Context, cfg config.Config) (*modules, error) {
	postgresClient, err := postgres.New(ctx, postgres.DefaultConfig(cfg.DatabaseURL))
	if err != nil {
		return nil, fmt.Errorf("initialize PostgreSQL: %w", err)
	}

	redisClient, err := redisintegration.New(ctx, redisintegration.DefaultConfig(cfg.RedisURL))
	if err != nil {
		postgresClient.Close()
		return nil, fmt.Errorf("initialize Redis: %w", err)
	}

	freeSwitch, err := freeswitch.New(
		freeswitch.DefaultConfig(cfg.FreeSWITCHESLAddress, cfg.FreeSWITCHESLPassword),
	)
	if err != nil {
		_ = redisClient.Close()
		postgresClient.Close()
		return nil, fmt.Errorf("initialize FreeSWITCH: %w", err)
	}
	if err := freeSwitch.Connect(ctx); err != nil {
		_ = freeSwitch.Close()
		_ = redisClient.Close()
		postgresClient.Close()
		return nil, fmt.Errorf("connect FreeSWITCH: %w", err)
	}

	closeDependencies := func() {
		_ = freeSwitch.Close()
		_ = redisClient.Close()
		postgresClient.Close()
	}

	credentialCipher, err := encryption.New(cfg.EncryptionKey)
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize trunk credential encryption: %w", err)
	}

	coturnClient, err := coturn.New(coturn.Config{
		AuthSecret: cfg.TURNAuthSecret,
		URLs:       cfg.TURNPublicURLs,
	})
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize Coturn integration: %w", err)
	}
	turnService, err := webrtc.NewService(coturnClient, redisClient)
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize TURN credentials: %w", err)
	}

	objectClient, err := minio.New(ctx, minio.DefaultConfig(
		cfg.Domain, cfg.MinIO.AccessKey, cfg.MinIO.SecretKey,
	))
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize recording object storage: %w", err)
	}

	queries := sqlc.New(postgresClient.Pool())
	identityModule := identity.New(
		queries,
		cfg.IsDevelopment(),
		cfg.Domain,
	)
	identityModule.Auth.Service.ConfigureDatabase(postgresClient.Pool())
	tenancyModule := tenancy.New(queries)
	trustedProxies, err := cfg.TrustedProxyPrefixes()
	if err != nil {
		closeDependencies()
		return nil, err
	}
	platformModule := platform.New(
		postgresClient.Pool(),
		queries,
		credentialCipher,
		trustedProxies,
	)
	stripeClient := stripe.New(cfg.Stripe.SecretKey, cfg.Stripe.WebhookSecret)
	commercialModule := commercial.New(queries, commercial.Dependencies{
		Stripe: stripeClient,
		StripePrices: map[string]string{
			"developer": cfg.Stripe.DeveloperPriceID,
			"pro":       cfg.Stripe.ProPriceID,
		},
	})
	platformModule.Retention.Service.SetPlanRetentionResolver(
		commercialModule.Limits.Service.RetentionDays,
	)

	recordingStorage := recordings.NewResolvedObjectStorage(
		objectClient,
		platformModule.Storage.Service,
	)

	metricsRegistry := metrics.New(redisClient)
	telephonyModule, err := telephony.New(telephony.Dependencies{
		DB:                postgresClient.Pool(),
		Queries:           queries,
		CallsController:   calling.NewController(freeSwitch),
		CallsChannelStore: calling.NewChannelStore(redisClient),
		CallsAdmission:    calling.NewAdmissionLimiter(redisClient),
		CredentialCipher:  credentialCipher,
		WebRTCService:     turnService,
		RecordingStorage:  recordingStorage,
		Metrics:           metricsRegistry,
	})
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize telephony: %w", err)
	}

	aiModule := ai.New(queries, ai.Dependencies{
		CredentialCipher: credentialCipher,
		PlatformCredentials: providers.PlatformCredentials{
			OpenAI:     cfg.OpenAIAPIKey,
			Gemini:     cfg.GeminiAPIKey,
			Deepgram:   cfg.DeepgramAPIKey,
			AssemblyAI: cfg.AssemblyAIAPIKey,
			Groq:       cfg.GroqAPIKey,
			Cartesia:   cfg.CartesiaAPIKey,
			ElevenLabs: cfg.ElevenLabsAPIKey,
		},
		Calls: telephonyModule.Calls.Service,
	})
	telephonyModule.Calls.Service.SetVoiceAgentReadinessCheck(
		aiModule.Agents.Service.RequireReady,
	)
	telephonyModule.Calls.Service.SetPlanConcurrencyResolver(
		commercialModule.Limits.Service.MaxConcurrentCalls,
	)

	resolver := authn.NewResolver(identityModule.Session.Service, tenancyModule.Credentials.Service)
	authMiddleware := middleware.NewAuthnMiddleware(resolver)
	organizationMiddleware := middleware.NewOrganizationMiddleware(queries)

	rateLimitStore, err := redisClient.NewRateLimitStore()
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize rate limit store: %w", err)
	}
	rateLimitMiddleware, err := middleware.NewRateLimitMiddleware(rateLimitStore)
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize rate limit middleware: %w", err)
	}

	return &modules{
		postgres:             postgresClient,
		redis:                redisClient,
		freeSwitch:           freeSwitch,
		identity:             identityModule,
		tenancy:              tenancyModule,
		platform:             platformModule,
		commercial:           commercialModule,
		ai:                   aiModule,
		telephony:            telephonyModule,
		authn:                authMiddleware,
		organizationsContext: organizationMiddleware,
		rateLimit:            rateLimitMiddleware,
		metrics:              metricsRegistry,
		queries:              queries,
		credentialCipher:     credentialCipher,
	}, nil
}

func (m *modules) close(logger *logging.Logger) {
	if m == nil {
		return
	}
	if m.freeSwitch != nil {
		if err := m.freeSwitch.Close(); err != nil {
			logger.Warn(context.Background(), "close FreeSWITCH", "error", err)
		}
	}
	if m.redis != nil {
		if err := m.redis.Close(); err != nil {
			logger.Warn(context.Background(), "close Redis", "error", err)
		}
	}
	if m.postgres != nil {
		m.postgres.Close()
	}
}
