package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/commercial/catalog"
	commercialstate "github.com/leamout/leamout/internal/commercial/state"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/internal/database"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/identity/session"
	"github.com/leamout/leamout/internal/integrations/redis"
	redisintegration "github.com/leamout/leamout/internal/integrations/redis"
	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/platform/encryption"
	"github.com/leamout/leamout/internal/platform/metrics"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/conferences"
	"github.com/leamout/leamout/internal/telecom/edge"
	"github.com/leamout/leamout/internal/telecom/numbers"
	"github.com/leamout/leamout/internal/telecom/realtime"
	"github.com/leamout/leamout/internal/telecom/recordings"
	"github.com/leamout/leamout/internal/telecom/routing"
	"github.com/leamout/leamout/internal/telecom/trunks"
	"github.com/leamout/leamout/internal/telecom/wholesale"
	"github.com/leamout/leamout/internal/tenancy/organization"
	"github.com/leamout/leamout/pkg/freeswitch"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Server struct {
	DB         *database.DB
	Router     *chi.Mux
	Modules    Modules
	FreeSWITCH *freeswitch.Client
	Redis      *redisintegration.Client
	Logger     *zap.Logger
	Metrics    *metrics.Registry
}

func New(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*Server, error) {
	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	redisClient := redis.NewClient(&redis.Options{Addr: cfg.RedisURL})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	freeSwitch, err := freeswitch.NewClient(cfg.FreeSWITCH.Address, cfg.FreeSWITCH.Password)
	if err != nil {
		_ = redisClient.Close()
		db.Close()
		return nil, err
	}

	credentialCipher, err := encryption.NewCipher(cfg.CarrierCredentialEncryptionKey)
	if err != nil {
		_ = freeSwitch.Close()
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize carrier credential encryption: %w", err)
	}

	redisIntegration := redisintegration.NewClient(redisClient)
	turnService := realtime.NewService(redisIntegration, cfg.TURNAuthSecret, cfg.TURNTTL, cfg.TURNPublicURLs)
	modules, err := NewModules(db.Pool, calls.NewFreeSWITCHController(freeSwitch), conferences.NewFreeSWITCHController(freeSwitch), credentialCipher, turnService, redisIntegration)
	if err != nil {
		_ = freeSwitch.Close()
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize modules: %w", err)
	}
	if err := configureManagedNumberAcquisition(cfg, modules.Numbers.Service); err != nil {
		_ = freeSwitch.Close()
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize managed number acquisition: %w", err)
	}
	if err := configureManagedSIP(cfg, modules.Trunks.Service, modules.CommercialState.Service); err != nil {
		_ = freeSwitch.Close()
		_ = redisClient.Close()
		db.Close()
		return nil, fmt.Errorf("initialize managed SIP: %w", err)
	}
	modules.Edge.Handler = edge.NewHandler(
		modules.Edge.Service,
		cfg.ManagedSIP.AdmissionSecret,
	)
	modules.Wholesale.Handler = wholesale.NewHandler(
		modules.Wholesale.Service,
		cfg.ManagedSIP.AdmissionSecret,
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
	organizationRepository := organization.NewRepository(queries)
	organizationService := organization.NewService(organizationRepository)

	routingRepository := routing.NewRepository(queries)
	routingResolver := routing.NewResolver(routingRepository)
	routingService := routing.NewService(routingResolver)

	trunksRepository := trunks.NewRepository(queries, credentialCipher)
	trunksService := trunks.NewService(trunksRepository)

	numbersRepository := numbers.NewRepository(queries)
	numbersService := numbers.NewService(numbersRepository)

	edgeRepository := edge.NewRepository(queries)
	edgeService := edge.NewService(edgeRepository, commercialStateService)

	wholesaleRepository := wholesale.NewRepository(queries)
	wholesaleService := wholesale.NewService(wholesaleRepository)

	callsRepository := calls.NewRepository(queries)
	callsService := calls.NewService(callsRepository, callsController, routingService)

	recordingsRepository := recordings.NewRepository(queries)
	recordingsService := recordings.NewService(recordingsRepository)

	conferenceRepository := conferences.NewRepository(queries)
	conferenceService := conferences.NewService(conferenceRepository, conferenceController)

	return Modules{
		Session: Module[session.Handler, session.Service]{
			Handler: session.NewHandler(sessionService),
			Service: sessionService,
		},
		Organizations: Module[organization.Handler, organization.Service]{
			Handler: organization.NewHandler(organizationService),
			Service: organizationService,
		},
		Routing: routingService,
		Trunks: Module[trunks.Handler, trunks.Service]{
			Handler: trunks.NewHandler(trunksService),
			Service: trunksService,
		},
		Numbers: Module[numbers.Handler, numbers.Service]{
			Handler: numbers.NewHandler(numbersService),
			Service: numbersService,
		},
		Edge: Module[edge.Handler, edge.Service]{
			Service: edgeService,
		},
		Wholesale: Module[wholesale.Handler, wholesale.Service]{
			Service: wholesaleService,
		},
		Calls: Module[calls.Handler, calls.Service]{
			Handler: calls.NewHandler(callsService),
			Service: callsService,
		},
		Recordings: Module[recordings.Handler, recordings.Service]{
			Handler: recordings.NewHandler(recordingsService),
			Service: recordingsService,
		},
		Conferences: Module[conferences.Handler, conferences.Service]{
			Handler: conferences.NewHandler(conferenceService),
			Service: conferenceService,
		},
	}, nil
}

var metricsRegistry = metrics.NewRegistry()

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}
