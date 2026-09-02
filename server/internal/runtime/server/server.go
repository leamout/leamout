package server

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/commercial/catalog"
	"github.com/leamout/leamout/internal/commercial/entitlements"
	"github.com/leamout/leamout/internal/commercial/licensing"
	commercialstate "github.com/leamout/leamout/internal/commercial/state"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/identity/auth"
	"github.com/leamout/leamout/internal/identity/session"
	"github.com/leamout/leamout/internal/identity/users"
	"github.com/leamout/leamout/internal/integrations/freeswitch"
	redisintegration "github.com/leamout/leamout/internal/integrations/redis"
	"github.com/leamout/leamout/internal/modules/audit"
	"github.com/leamout/leamout/internal/modules/webhooks"
	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/platform/logging"
	"github.com/leamout/leamout/internal/platform/metrics"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/internal/security/authn"
	"github.com/leamout/leamout/internal/security/secrets"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/carriers"
	"github.com/leamout/leamout/internal/telecom/conferences"
	"github.com/leamout/leamout/internal/telecom/numbers"
	"github.com/leamout/leamout/internal/telecom/realtime"
	"github.com/leamout/leamout/internal/telecom/recordings"
	"github.com/leamout/leamout/internal/telecom/routing"
	"github.com/leamout/leamout/internal/telecom/sip_domains"
	"github.com/leamout/leamout/internal/telecom/subscribers"
	"github.com/leamout/leamout/internal/telecom/trunks"
	"github.com/leamout/leamout/internal/telecom/voice"
	"github.com/leamout/leamout/internal/tenancy/credentials"
	"github.com/leamout/leamout/internal/tenancy/members"
	"github.com/leamout/leamout/internal/tenancy/organization"
)

type Server struct {
	DB         *pgxpool.Pool
	Router     *chi.Mux
	Modules    Modules
	FreeSWITCH freeswitch.MediaController
	Redis      *redisintegration.Client

	Logger  *logging.Logger
	Metrics *metrics.Registry
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

	freeSwitch, err := freeswitch.New(
		freeswitch.DefaultConfig(cfg.FreeSWITCHESLAddress, cfg.FreeSWITCHESLPassword),
	)
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

	callsController := calls.NewFreeSWITCHController(freeSwitch)
	conferenceController := conferences.NewFreeSWITCHController(freeSwitch)

	logger := logging.New()
	metricsRegistry := metrics.New(redisClient)

	credentialCipher, err := secrets.New(cfg.CarrierCredentialKey)
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
	modules, err := NewModules(db, callsController, conferenceController, credentialCipher, turnService, redisClient)
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
	credentialCipher *secrets.Cipher,
	turnService *realtime.Service,
	redisClient *redisintegration.Client,
) (Modules, error) {
	queries := sqlc.New(db)
	catalogRepository := catalog.NewRepository(db)
	catalogService := catalog.NewService(catalogRepository)
	subscriptionsRepository := subscriptions.NewRepository(db)
	subscriptionsService := subscriptions.NewService(subscriptionsRepository, catalogService)
	entitlementsRepository := entitlements.NewRepository(db)
	entitlementsService := entitlements.NewService(entitlementsRepository, subscriptionsService)
	commercialStateService := commercialstate.NewService(subscriptionsService, entitlementsService)
	licensingRepository := licensing.NewRepository(db)
	licensingService := licensing.NewService(licensingRepository, commercialStateService)

	sessionRepository := session.NewRepository(queries)
	sessionService := session.NewService(sessionRepository)

	authRepository := auth.NewRepository(queries)
	authService := auth.NewService(authRepository)

	usersRepository := users.NewRepository(queries)
	usersService := users.NewService(usersRepository)

	organizationRepository := organization.NewRepository(queries)
	organizationService := organization.NewService(organizationRepository)

	membersRepository := members.NewRepository(queries)
	membersService := members.NewService(membersRepository)

	credentialsRepository := credentials.NewRepository(queries)
	credentialsService := credentials.NewService(credentialsRepository)

	voiceRepository := voice.NewRepository(queries)
	voiceService := voice.NewService(voiceRepository)

	routingRepository := routing.NewRepository(queries)
	routeResolver := routing.NewResolver(routingRepository)
	telecomMetrics := metrics.New(redisClient)
	routeResolver.SetMetrics(telecomMetrics)
	routingService := routing.NewService(routeResolver)

	callsRepository := calls.NewRepository(db)
	callAdmission, err := calls.NewAdmissionController(redisClient, callsRepository)
	if err != nil {
		return Modules{}, err
	}
	callsService := calls.NewService(callsRepository, callsController, routingService, callAdmission)
	callsService.SetMetrics(telecomMetrics)

	recordingsRepository := recordings.NewRepository(db)
	recordingsService := recordings.NewService(recordingsRepository, nil)

	conferencesRepository := conferences.NewRepository(db)
	conferencesService := conferences.NewService(conferencesRepository, conferenceController)

	subscribersRepository := subscribers.NewRepository(queries)
	subscribersService := subscribers.NewService(subscribersRepository)

	numbersRepository := numbers.NewRepository(db)
	numbersService := numbers.NewService(numbersRepository)

	sipDomainsRepository := sip_domains.NewRepository(queries)
	sipDomainsService := sip_domains.NewService(sipDomainsRepository)
	carriersRepository := carriers.NewRepository(db)
	carriersService := carriers.NewService(carriersRepository, credentialCipher)

	trunksRepository := trunks.NewRepository(queries)
	trunksService := trunks.NewService(trunksRepository, db)

	webhooksRepository := webhooks.NewRepository(queries)
	webhooksService := webhooks.NewService(webhooksRepository)
	auditRepository := audit.NewRepository(db)
	auditService := audit.NewService(auditRepository)

	resolver := authn.NewResolver(
		sessionService,
		credentialsService,
	)

	authMiddleware := middleware.NewAuthnMiddleware(resolver)
	organizationMiddleware := middleware.NewOrganizationMiddleware()

	return Modules{
		Catalog: CatalogModule{
			Repository: catalogRepository,
			Service:    catalogService,
			Handler:    catalog.NewHandler(catalogService),
		},
		Licensing: LicensingModule{
			Repository: licensingRepository,
			Service:    licensingService,
			Handler:    licensing.NewHandler(licensingService),
		},
		Subscriptions: SubscriptionsModule{
			Repository: subscriptionsRepository,
			Service:    subscriptionsService,
			Handler:    subscriptions.NewHandler(subscriptionsService),
		},
		Auth: AuthModule{
			Repository: authRepository,
			Service:    authService,
			Handler:    auth.NewHandler(authService, sessionService),
		},
		Session: SessionModule{
			Repository: sessionRepository,
			Service:    sessionService,
			Handler:    session.NewHandler(sessionService),
		},
		Users: UsersModule{
			Repository: usersRepository,
			Service:    usersService,
			Handler:    users.NewHandler(usersService),
		},
		Organizations: OrganizationModule{
			Repository: organizationRepository,
			Service:    organizationService,
			Handler:    organization.NewHandler(organizationService),
		},
		Members: MembersModule{
			Repository: membersRepository,
			Service:    membersService,
			Handler:    members.NewHandler(membersService),
		},
		Credentials: CredentialsModule{
			Repository: credentialsRepository,
			Service:    credentialsService,
			Handler:    credentials.NewHandler(credentialsService),
		},
		Voice: VoiceModule{
			Repository: voiceRepository,
			Service:    voiceService,
			Handler:    voice.NewHandler(voiceService),
		},
		Calls: CallsModule{
			Repository: callsRepository,
			Service:    callsService,
			Handler:    calls.NewHandler(callsService),
		},
		Recordings: RecordingsModule{
			Repository: recordingsRepository,
			Service:    recordingsService,
			Handler:    recordings.NewHandler(recordingsService),
		},
		Subscribers: SubscribersModule{
			Repository: subscribersRepository,
			Service:    subscribersService,
			Handler:    subscribers.NewHandler(subscribersService),
		},
		Numbers: NumbersModule{
			Repository: numbersRepository,
			Service:    numbersService,
			Handler:    numbers.NewHandler(numbersService),
		},
		SIPDomains: SIPDomainsModule{
			Repository: sipDomainsRepository,
			Service:    sipDomainsService,
			Handler:    sip_domains.NewHandler(sipDomainsService),
		},
		Carriers: CarriersModule{
			Repository: carriersRepository,
			Service:    carriersService,
			Handler:    carriers.NewHandler(carriersService),
		},
		Trunks: TrunksModule{
			Repository: trunksRepository,
			Service:    trunksService,
			Handler:    trunks.NewHandler(trunksService),
		},
		Webhooks: WebhooksModule{
			Repository: webhooksRepository,
			Service:    webhooksService,
			Handler:    webhooks.NewHandler(webhooksService),
		},
		Audit: AuditModule{
			Repository: auditRepository,
			Service:    auditService,
			Handler:    audit.NewHandler(auditService),
		},
		Conferences: ConferencesModule{
			Repository: conferencesRepository,
			Service:    conferencesService,
			Handler:    conferences.NewHandler(conferencesService),
		},
		Realtime: RealtimeModule{
			Service: turnService,
			Handler: realtime.NewHandler(turnService),
		},
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
