package server

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/identity/auth"
	"github.com/leamout/leamout/internal/identity/session"
	"github.com/leamout/leamout/internal/identity/users"
	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/platform/logging"
	"github.com/leamout/leamout/internal/platform/metrics"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/internal/security/authn"
	"github.com/leamout/leamout/internal/tenancy/credentials"
	"github.com/leamout/leamout/internal/tenancy/members"
	"github.com/leamout/leamout/internal/tenancy/organization"
)

type Server struct {
	DB      *pgxpool.Pool
	Router  *chi.Mux
	Modules Modules

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

	logger := logging.New()
	metricsRegistry := metrics.New()

	modules, err := NewModules(db)
	if err != nil {
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

	RegisterRoutes(router, modules)

	return &Server{
		DB:      db,
		Router:  router,
		Modules: modules,
		Logger:  logger,
		Metrics: metricsRegistry,
	}, nil
}

func NewModules(db *pgxpool.Pool) (Modules, error) {
	queries := sqlc.New(db)

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

	resolver := authn.NewResolver(
		sessionService,
		nil,
	)

	authMiddleware := middleware.NewAuthnMiddleware(resolver)

	return Modules{
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
		Authn: authMiddleware,
	}, nil
}

func (s *Server) Close() {
	if s.DB != nil {
		s.DB.Close()
	}
}
