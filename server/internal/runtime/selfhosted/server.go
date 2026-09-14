package server

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/identity"
	"github.com/leamout/leamout/internal/integrations/freeswitch"
	redisintegration "github.com/leamout/leamout/internal/integrations/redis"
	sharedmodules "github.com/leamout/leamout/internal/modules"
	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/platform/logging"
	"github.com/leamout/leamout/internal/platform/metrics"
	"github.com/leamout/leamout/internal/platform/middleware"
	"github.com/leamout/leamout/internal/security/authn"
	"github.com/leamout/leamout/internal/security/encryption"
	"github.com/leamout/leamout/internal/telecom"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/conferences"
	"github.com/leamout/leamout/internal/telecom/realtime"
	"github.com/leamout/leamout/internal/tenancy"
)

type Server struct {
	DB         *pgxpool.Pool
	Router     *chi.Mux
	Modules    Modules
	FreeSWITCH freeswitch.MediaController
	Redis      *redisintegration.Client
	Logger     *logging.Logger
	Metrics    *metrics.Registry
}

func New(ctx context.Context, cfg config.Config) (*Server, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	redisClient, err := redisintegration.New(ctx, redisintegration.DefaultConfig(cfg.RedisURL))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("connect Redis: %w", err)
	}

	freeSwitch, err := freeswitch.New(freeswitch.DefaultConfig(cfg.FreeSWITCHESLAddress, cfg.FreeSWITCHESLPassword))
	if err != nil {
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize FreeSWITCH client: %w", err)
	}
	if err := freeSwitch.Connect(ctx); err != nil {
		_ = freeSwitch.Close()
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("connect FreeSWITCH: %w", err)
	}

	logger := logging.New()
	metricsRegistry := metrics.New(redisClient)
	credentialCipher, err := encryption.New(cfg.CarrierCredentialKey)
	if err != nil {
		_ = freeSwitch.Close()
		_ = redisClient.Close()
		db.Close()
		return nil, err
	}
	turnService, err := realtime.NewService(realtime.Config{
		AuthSecret: cfg.TURNAuthSecret,
		URLs:       cfg.TURNPublicURLs,
	}, redisClient)
	if err != nil {
		_ = freeSwitch.Close()
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize TURN credentials: %w", err)
	}

	modules, err := NewModules(
		db,
		calls.NewFreeSWITCHController(freeSwitch),
		conferences.NewFreeSWITCHController(freeSwitch),
		credentialCipher,
		turnService,
		redisClient,
	)
	if err != nil {
		_ = freeSwitch.Close()
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize modules: %w", err)
	}

	router := chi.NewRouter()
	router.Use(
		middleware.Recovery,
		middleware.Tracing(),
		middleware.Request(),
		middleware.Logging(logger),
		middleware.Metrics(metricsRegistry),
		middleware.Secure,
		middleware.CORS(cfg.CORSOrigins, cfg.IsDevelopment()),
	)
	RegisterHealthRoutes(router, db, redisClient, freeSwitch)
	router.Handle("/metrics", metrics.Handler(metricsRegistry))
	RegisterRoutes(router, modules)

	return &Server{
		DB:         db,
		Router:     router,
		Modules:    modules,
		FreeSWITCH: freeSwitch,
		Redis:      redisClient,
		Logger:     logger,
		Metrics:    metricsRegistry,
	}, nil
}

func NewModules(
	db *pgxpool.Pool,
	callsController calls.Controller,
	conferenceController conferences.Controller,
	credentialCipher *encryption.Cipher,
	turnService *realtime.Service,
	redisClient *redisintegration.Client,
) (Modules, error) {
	queries := sqlc.New(db)
	identityModule := identity.New(queries)
	tenancyModule := tenancy.New(queries)
	sharedModule := sharedmodules.New(db, queries)
	telecomModule, err := telecom.New(telecom.Dependencies{
		DB:                   db,
		Queries:              queries,
		Redis:                redisClient,
		CallsController:      callsController,
		ConferenceController: conferenceController,
		CredentialCipher:     credentialCipher,
		RealtimeService:      turnService,
	})
	if err != nil {
		return Modules{}, err
	}

	resolver := authn.NewResolver(identityModule.Session.Service, tenancyModule.Credentials.Service)
	authMiddleware := middleware.NewAuthnMiddleware(resolver)
	organizationMiddleware := middleware.NewOrganizationMiddleware(queries)
	rateLimitStore, err := redisClient.NewRateLimitStore()
	if err != nil {
		return Modules{}, fmt.Errorf("initialize rate limit store: %w", err)
	}
	rateLimitMiddleware, err := middleware.NewRateLimitMiddleware(rateLimitStore)
	if err != nil {
		return Modules{}, fmt.Errorf("initialize rate limit middleware: %w", err)
	}

	return Modules{
		Identity:             identityModule,
		Tenancy:              tenancyModule,
		Shared:               sharedModule,
		Telecom:              telecomModule,
		RateLimit:            rateLimitMiddleware,
		Authn:                authMiddleware,
		OrganizationsContext: organizationMiddleware,
	}, nil
}

func (s *Server) Close() {
	if s.FreeSWITCH != nil {
		_ = s.FreeSWITCH.Close()
	}
	if s.Redis != nil {
		_ = s.Redis.Close()
	}
	if s.DB != nil {
		s.DB.Close()
	}
}
