package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/commercial"
	"github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/identity"
	"github.com/leamout/leamout/internal/integrations/carriers/didww"
	"github.com/leamout/leamout/internal/integrations/freeswitch"
	"github.com/leamout/leamout/internal/integrations/payments/paystack"
	"github.com/leamout/leamout/internal/integrations/payments/stripe"
	redisintegration "github.com/leamout/leamout/internal/integrations/redis"
	sharedmodules "github.com/leamout/leamout/internal/modules"
	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/platform/logging"
	"github.com/leamout/leamout/internal/platform/metrics"
	"github.com/leamout/leamout/internal/platform/middleware"
	providerdiagnostics "github.com/leamout/leamout/internal/platform/provider_diagnostics"
	"github.com/leamout/leamout/internal/security/authn"
	"github.com/leamout/leamout/internal/security/encryption"
	"github.com/leamout/leamout/internal/telecom"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/conferences"
	"github.com/leamout/leamout/internal/telecom/edge"
	"github.com/leamout/leamout/internal/telecom/numbers"
	"github.com/leamout/leamout/internal/telecom/realtime"
	"github.com/leamout/leamout/internal/telecom/wholesale"
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

	if err := configurePaymentProviders(cfg, modules.Commercial.Billing.Payments.Providers); err != nil {
		_ = freeSwitch.Close()
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize payment providers: %w", err)
	}
	if err := configureManagedNumberAcquisition(cfg, modules.Telecom.Numbers.Service); err != nil {
		_ = freeSwitch.Close()
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize managed number acquisition: %w", err)
	}

	modules.Edge.Handler = edge.NewHandler(modules.Edge.Service, cfg.ManagedSIP.AdmissionSecret)
	modules.Wholesale.Handler = wholesale.NewHandler(modules.Wholesale.Service, cfg.ManagedSIP.AdmissionSecret)
	modules.ProviderDiagnostics.Handler = providerdiagnostics.NewHandler(
		modules.ProviderDiagnostics.Service,
		cfg.OperatorAPISecret,
	)

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

	commercialModule := commercial.NewCloud(db)
	telecomModule.Numbers.Service.SetManagedPurchaseAuthority(commercialModule.Wallets.Service)
	edgeRepository := edge.NewRepository(db)
	edgeService := edge.NewService(edgeRepository)
	wholesaleRepository := wholesale.NewRepository(db)
	wholesaleService := wholesale.NewService(wholesaleRepository)
	providerDiagnosticsRepository := providerdiagnostics.NewRepository(queries)
	providerDiagnosticsService := providerdiagnostics.NewService(providerDiagnosticsRepository)

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
		Commercial:           commercialModule,
		Identity:             identityModule,
		Tenancy:              tenancyModule,
		Shared:               sharedModule,
		Telecom:              telecomModule,
		Edge:                 EdgeModule{Repository: edgeRepository, Service: edgeService},
		Wholesale:            WholesaleModule{Repository: wholesaleRepository, Service: wholesaleService},
		ProviderDiagnostics:  ProviderDiagnosticsModule{Repository: providerDiagnosticsRepository, Service: providerDiagnosticsService},
		RateLimit:            rateLimitMiddleware,
		Authn:                authMiddleware,
		OrganizationsContext: organizationMiddleware,
	}, nil
}

func configureManagedNumberAcquisition(cfg config.Config, service *numbers.Service) error {
	if strings.TrimSpace(cfg.DIDWW.APIKey) == "" {
		return nil
	}
	client, err := didww.NewClient(didww.Config{
		BaseURL: cfg.DIDWW.APIBaseURL,
		APIKey:  cfg.DIDWW.APIKey,
	})
	if err != nil {
		return err
	}
	service.SetManagedAcquisition(client)
	return nil
}

func configurePaymentProviders(cfg config.Config, providers *payments.ProviderRegistry) error {
	if cfg.Stripe.SecretKey != "" {
		if cfg.Stripe.WebhookSecret == "" {
			return fmt.Errorf("stripe webhook secret is required when Stripe is enabled")
		}
		client, err := stripe.NewClient(stripe.Config{
			BaseURL:       cfg.Stripe.APIBaseURL,
			SecretKey:     cfg.Stripe.SecretKey,
			WebhookSecret: cfg.Stripe.WebhookSecret,
		})
		if err != nil {
			return err
		}
		providers.Set("stripe", client)
	}
	if cfg.Paystack.SecretKey != "" {
		client, err := paystack.NewClient(paystack.Config{
			BaseURL:   cfg.Paystack.APIBaseURL,
			SecretKey: cfg.Paystack.SecretKey,
		})
		if err != nil {
			return err
		}
		providers.Set("paystack", client)
	}
	return nil
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
