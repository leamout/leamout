package backoffice

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/internal/database/sqlc"
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

	modules := newModules(sqlc.New(db))
	return newServer(db, modules), nil
}

func newServer(db *pgxpool.Pool, modules Modules) *Server {
	router := chi.NewRouter()
	router.Use(securityHeaders)
	registerRoutes(router, modules)
	return &Server{DB: db, Router: router, Modules: modules}
}

func (s *Server) Close() {
	if s != nil && s.DB != nil {
		s.DB.Close()
	}
}
