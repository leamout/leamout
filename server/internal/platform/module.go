package platform

import (
	"net/netip"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/platform/audit"
	"github.com/leamout/leamout/server/internal/platform/idempotency"
	"github.com/leamout/leamout/server/internal/platform/middleware"
	"github.com/leamout/leamout/server/internal/platform/networking"
	"github.com/leamout/leamout/server/internal/platform/retention"
	"github.com/leamout/leamout/server/internal/platform/storage"
	"github.com/leamout/leamout/server/internal/platform/webhooks"
	"github.com/leamout/leamout/server/internal/security/encryption"
)

type Module struct {
	Audit       AuditModule
	Idempotency IdempotencyModule
	Networking  NetworkingModule
	Storage     StorageModule
	Retention   RetentionModule
	Webhooks    WebhooksModule
}

type AuditModule struct {
	Repository *audit.Repository
	Service    *audit.Service
	Handler    *audit.Handler
}

type IdempotencyModule struct {
	Repository *idempotency.Repository
	Service    *idempotency.Service
	Middleware *middleware.IdempotencyMiddleware
}

type NetworkingModule struct {
	Repository *networking.Repository
	Service    *networking.Service
	Handler    *networking.Handler
	Middleware *middleware.NetworkingMiddleware
}

type StorageModule struct {
	Repository *storage.Repository
	Service    *storage.Service
	Handler    *storage.Handler
}

type RetentionModule struct {
	Repository *retention.Repository
	Service    *retention.Service
	Handler    *retention.Handler
}

type WebhooksModule struct {
	Repository *webhooks.Repository
	Service    *webhooks.Service
	Handler    *webhooks.Handler
}

func New(
	db *pgxpool.Pool,
	queries *sqlc.Queries,
	credentialCipher *encryption.Cipher,
	trustedProxies []netip.Prefix,
) *Module {
	auditRepository := audit.NewRepository(db)
	auditService := audit.NewService(auditRepository)

	idempotencyRepository := idempotency.NewRepository(queries)
	idempotencyService := idempotency.NewService(
		idempotencyRepository,
		idempotency.DefaultConfig(),
	)

	networkingRepository := networking.NewRepository(queries)
	networkingService := networking.NewService(networkingRepository)

	storageRepository := storage.NewRepository(queries)
	storageService := storage.NewService(storageRepository, credentialCipher)

	retentionRepository := retention.NewRepository(queries)
	retentionService := retention.NewService(retentionRepository)

	webhooksRepository := webhooks.NewRepository(queries)
	webhooksService := webhooks.NewService(webhooksRepository)

	return &Module{
		Audit: AuditModule{
			Repository: auditRepository,
			Service:    auditService,
			Handler:    audit.NewHandler(auditService),
		},
		Idempotency: IdempotencyModule{
			Repository: idempotencyRepository,
			Service:    idempotencyService,
			Middleware: middleware.NewIdempotencyMiddleware(idempotencyService),
		},
		Networking: NetworkingModule{
			Repository: networkingRepository,
			Service:    networkingService,
			Handler:    networking.NewHandler(networkingService),
			Middleware: middleware.NewNetworkingMiddleware(
				networkingService,
				trustedProxies,
			),
		},
		Storage: StorageModule{
			Repository: storageRepository,
			Service:    storageService,
			Handler:    storage.NewHandler(storageService),
		},
		Retention: RetentionModule{
			Repository: retentionRepository,
			Service:    retentionService,
			Handler:    retention.NewHandler(retentionService),
		},
		Webhooks: WebhooksModule{
			Repository: webhooksRepository,
			Service:    webhooksService,
			Handler:    webhooks.NewHandler(webhooksService),
		},
	}
}
