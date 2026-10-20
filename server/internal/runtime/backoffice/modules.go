package backoffice

import (
	"context"
	"fmt"

	"github.com/leamout/leamout/server/internal/backoffice/organizations"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/identity/session"
	"github.com/leamout/leamout/server/internal/integrations/keycloak"
	"github.com/leamout/leamout/server/internal/integrations/postgres"
	"github.com/leamout/leamout/server/internal/observability/logging"
	"github.com/leamout/leamout/server/internal/platform/config"
	"github.com/leamout/leamout/server/internal/platform/middleware"
	"github.com/leamout/leamout/server/internal/security/authn"
	"github.com/leamout/leamout/server/internal/security/encryption"
)

type modules struct {
	postgres      *postgres.Client
	queries       *sqlc.Queries
	authn         *middleware.AuthnMiddleware
	organizations *organizations.Handler
}

func newModules(ctx context.Context, cfg config.BackofficeConfig, logger *logging.Logger) (*modules, error) {
	cipher, err := encryption.New(cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("initialize session encryption: %w", err)
	}
	client, err := keycloak.New(ctx, cfg.Keycloak, cfg.FrontendURL+"/api/v1/auth/callback", cfg.IsDevelopment())
	if err != nil {
		return nil, fmt.Errorf("initialize Keycloak: %w", err)
	}
	db, err := postgres.New(ctx, postgres.DefaultConfig(cfg.DatabaseURL))
	if err != nil {
		return nil, fmt.Errorf("initialize PostgreSQL: %w", err)
	}
	queries := sqlc.New(db.Pool())
	sessions := session.NewService(session.NewRepository(queries), db.Pool(), client, cipher)
	resolver := authn.NewResolver(sessions, nil)
	return &modules{
		postgres:      db,
		queries:       queries,
		authn:         middleware.NewAuthnMiddleware(resolver),
		organizations: organizations.NewHandler(organizations.NewRepository(queries), logger),
	}, nil
}
