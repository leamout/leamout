package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/leamout/leamout/server/internal/ai"
	"github.com/leamout/leamout/server/internal/commercial"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/identity"
	"github.com/leamout/leamout/server/internal/integrations/coturn"
	"github.com/leamout/leamout/server/internal/integrations/freeswitch"
	"github.com/leamout/leamout/server/internal/integrations/keycloak"
	"github.com/leamout/leamout/server/internal/integrations/payments/polar"
	"github.com/leamout/leamout/server/internal/integrations/postgres"
	redisintegration "github.com/leamout/leamout/server/internal/integrations/redis"
	"github.com/leamout/leamout/server/internal/observability/logging"
	"github.com/leamout/leamout/server/internal/observability/metrics"
	"github.com/leamout/leamout/server/internal/platform"
	"github.com/leamout/leamout/server/internal/platform/config"
	"github.com/leamout/leamout/server/internal/platform/email"
	"github.com/leamout/leamout/server/internal/platform/middleware"
	"github.com/leamout/leamout/server/internal/runtime/calling"
	"github.com/leamout/leamout/server/internal/security/authn"
	"github.com/leamout/leamout/server/internal/security/encryption"
	"github.com/leamout/leamout/server/internal/telephony"
	"github.com/leamout/leamout/server/internal/telephony/recordings"
	"github.com/leamout/leamout/server/internal/telephony/webrtc"
	"github.com/leamout/leamout/server/pkg/apperror"
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

	queries := sqlc.New(postgresClient.Pool())
	keycloakClient, err := keycloak.New(ctx, cfg.Keycloak, cfg.FrontendURL+"/api/v1/auth/callback", cfg.IsDevelopment())
	if err != nil {
		closeDependencies()
		return nil, fmt.Errorf("initialize Keycloak: %w", err)
	}
	identityModule := identity.New(queries, cfg.IsDevelopment(), postgresClient.Pool(), keycloakClient, credentialCipher, cfg.FrontendURL)
	tenancyModule := tenancy.New(queries)
	invitationURL := cfg.FrontendURL + "/invitations/accept"
	tenancyModule.ConfigureEmail(postgresClient.Pool(), email.NewService(credentialCipher), invitationURL)
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
	commercialModule := commercial.New(queries, commercial.Dependencies{
		Database:    postgresClient.Pool(),
		Polar:       polar.New(cfg.Polar.AccessToken, cfg.Polar.WebhookSecret, cfg.Polar.Environment),
		FrontendURL: cfg.FrontendURL,
		PolarProducts: map[string]string{
			"developer": cfg.Polar.DeveloperProductID,
			"pro":       cfg.Polar.ProProductID,
			"scale":     cfg.Polar.ScaleProductID,
		},
	})
	platformModule.Retention.Service.SetPlanRetentionResolver(
		commercialModule.Limits.Service.RetentionDays,
	)

	recordingStorage := recordings.NewObjectStorage(platformModule.Storage.Service)

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
		Calls:            telephonyModule.Calls.Service,
		Database:         postgresClient.Pool(),
		PlanAgentLimit:   commercialModule.Limits.Service.MaxVoiceAgents,
	})
	telephonyModule.Calls.Service.SetVoiceAgentReadinessCheck(
		aiModule.Agents.Service.RequireReady,
	)
	telephonyModule.Calls.Service.SetPlanConcurrencyResolver(
		commercialModule.Limits.Service.MaxConcurrentCalls,
	)
	telephonyModule.Calls.Service.SetRuntimeAllowanceCheck(
		commercialModule.Limits.Service.RequireRuntimeAllowance,
	)
	telephonyModule.Calls.Service.SetRecordingStorageCheck(
		func(ctx context.Context, organizationID uuid.UUID) error {
			_, err := platformModule.Storage.Service.ResolveRecording(ctx, organizationID)
			if errors.Is(err, pgx.ErrNoRows) {
				return apperror.NewConflict(
					"configure customer-owned S3-compatible storage before recording",
				)
			}
			if err != nil {
				return apperror.NewServiceUnavailable("recording storage is unavailable", err)
			}
			return nil
		},
	)
	telephonyModule.Trunks.Service.SetPlanTrunkLimit(
		commercialModule.Limits.Service.MaxSIPTrunks,
	)
	aiModule.Orchestration.SetRuntimeAllowanceCheck(
		commercialModule.Limits.Service.RequireRuntimeAllowance,
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
