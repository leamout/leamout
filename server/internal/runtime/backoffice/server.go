package backoffice

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/sqlc"
	"github.com/leamout/leamout/internal/identity/session"
	"github.com/leamout/leamout/internal/platform/config"
	platformmiddleware "github.com/leamout/leamout/internal/platform/middleware"
	"github.com/leamout/leamout/internal/security/authn"
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
	resolver := authn.NewResolver(sessionService, nil)
	authMiddleware := platformmiddleware.NewAuthnMiddleware(resolver)

	return newServer(
		db,
		modules,
		authMiddleware.RequireSession,
		requirePlatformAdmin(queries),
		protectUnsafeRequests,
	), nil
}

func newServer(
	db *pgxpool.Pool,
	modules Modules,
	access ...accessMiddleware,
) *Server {
	router := chi.NewRouter()
	router.Use(securityHeaders)
	registerRoutes(router, modules, access...)
	return &Server{DB: db, Router: router, Modules: modules}
}

func (s *Server) Close() {
	if s != nil && s.DB != nil {
		s.DB.Close()
	}
}
