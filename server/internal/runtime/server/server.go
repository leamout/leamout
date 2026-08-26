package server

import (
	"context"
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/identity/auth"
	"github.com/leamout/leamout/internal/identity/session"
	"github.com/leamout/leamout/internal/platform/config"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/internal/security/authn"
)

type Server struct {
	DB      *pgxpool.Pool
	Router  *chi.Mux
	Modules Modules
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

	modules, err := NewModules(db)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize modules: %w", err)
	}

	router := chi.NewRouter()
	RegisterRoutes(router, modules)

	return &Server{
		DB:      db,
		Router:  router,
		Modules: modules,
	}, nil
}

func NewModules(db *pgxpool.Pool) (Modules, error) {
	queries := sqlc.New(db)

	sessionRepository := session.NewRepository(queries)
	sessionService := session.NewService(sessionRepository)

	authRepository := auth.NewRepository(queries)
	authService := auth.NewService(authRepository)

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
		Authn: authMiddleware,
	}, nil
}

func (s *Server) Close() {
	if s.DB != nil {
		s.DB.Close()
	}
}
