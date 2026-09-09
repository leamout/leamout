package backoffice

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	backofficeauth "github.com/leamout/leamout/internal/backoffice/auth"
	"github.com/leamout/leamout/internal/database/sqlc"
	identityauth "github.com/leamout/leamout/internal/identity/auth"
	"github.com/leamout/leamout/internal/identity/session"
	"github.com/leamout/leamout/internal/platform/config"
)

type Server struct {
	DB      *pgxpool.Pool
	Router  *chi.Mux
	Modules Modules
}

func New(ctx context.Context, cfg config.BackofficeConfig) (*Server, error) {
	if strings.TrimSpace(cfg.DatabaseURL) == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	db, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect backoffice database: %w", err)
	}
	if err := db.Ping(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping backoffice database: %w", err)
	}

	queries := sqlc.New(db)
	modules := newModules(queries)

	sessionRepository := session.NewRepository(queries)
	sessionService := session.NewService(sessionRepository)
	identityAuthRepository := identityauth.NewRepository(queries)
	identityAuthService := identityauth.NewService(identityAuthRepository)
	backofficeAuthRepository := backofficeauth.NewRepository(queries)
	backofficeAuthService := backofficeauth.NewService(
		identityAuthService,
		sessionService,
		backofficeAuthRepository,
	)
	backofficeAuthHandler := backofficeauth.NewHandler(backofficeAuthService)

	return newServer(
		db,
		modules,
		backofficeAuthHandler,
		backofficeAuthService,
	), nil
}

func newServer(
	db *pgxpool.Pool,
	modules Modules,
	authHandler *backofficeauth.Handler,
	authentication authenticator,
) *Server {
	router := chi.NewRouter()
	router.Use(securityHeaders)
	registerRoutes(router, modules, authHandler, authentication)
	return &Server{DB: db, Router: router, Modules: modules}
}

func (s *Server) Close() {
	if s != nil && s.DB != nil {
		s.DB.Close()
	}
}
