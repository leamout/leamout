package backoffice

import (
	"context"
	"fmt"

	"github.com/leamout/leamout/server/internal/backoffice/agents"
	backofficeauth "github.com/leamout/leamout/server/internal/backoffice/auth"
	"github.com/leamout/leamout/server/internal/backoffice/organizations"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	identityauth "github.com/leamout/leamout/server/internal/identity/auth"
	"github.com/leamout/leamout/server/internal/identity/session"
	"github.com/leamout/leamout/server/internal/integrations/kinde"
	"github.com/leamout/leamout/server/internal/integrations/postgres"
	"github.com/leamout/leamout/server/internal/observability/logging"
	"github.com/leamout/leamout/server/internal/platform/config"
	"github.com/leamout/leamout/server/internal/platform/middleware"
	"github.com/leamout/leamout/server/internal/security/authn"
	"github.com/leamout/leamout/server/internal/security/encryption"
)

type modules struct {
	authentication *backofficeauth.Handler
	postgres       *postgres.Client
	queries        *sqlc.Queries
	authn          *middleware.AuthnMiddleware
	organizations  *organizations.Handler
	agents         *agents.Handler
}

func newModules(ctx context.Context, cfg config.Config, logger *logging.Logger) (*modules, error) {
	cipher, err := encryption.New(cfg.EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("initialize session encryption: %w", err)
	}
	client, err := kinde.New(ctx, cfg.BackofficeKinde, cfg.BackofficeURL+"/auth/callback", cfg.IsDevelopment())
	if err != nil {
		return nil, fmt.Errorf("initialize Kinde: %w", err)
	}
	db, err := postgres.New(ctx, postgres.DefaultConfig(cfg.DatabaseURL))
	if err != nil {
		return nil, fmt.Errorf("initialize PostgreSQL: %w", err)
	}
	queries := sqlc.New(db.Pool())
	sessions := session.NewService(session.NewRepository(queries), db.Pool(), client, cipher)
	resolver := authn.NewResolver(sessions, nil)
	authentication := identityauth.NewHandler(queries, sessions, client, cipher, cfg.BackofficeURL, cfg.IsDevelopment())
	return &modules{
		authentication: backofficeauth.NewHandler(authentication),
		postgres:       db,
		queries:        queries,
		authn:          middleware.NewAuthnMiddleware(resolver),
		organizations:  organizations.NewHandler(organizations.NewRepository(queries), logger),
		agents:         agents.NewHandler(agents.NewRepository(queries), logger),
	}, nil
}
