package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/leamout/leamout/server/internal/ai"
	"github.com/leamout/leamout/server/internal/identity"
	"github.com/leamout/leamout/server/internal/platform"
	"github.com/leamout/leamout/server/internal/platform/config"
	"github.com/leamout/leamout/server/internal/platform/logging"
	"github.com/leamout/leamout/server/internal/platform/metrics"
	"github.com/leamout/leamout/server/internal/platform/middleware"
	"github.com/leamout/leamout/server/internal/telephony"
	"github.com/leamout/leamout/server/internal/tenancy"
)

func newRouter(cfg config.Config, logger *logging.Logger, modules *modules) *chi.Mux {
	router := chi.NewRouter()
	router.Use(
		middleware.Recovery,
		middleware.Tracing(),
		middleware.Request(),
		middleware.Logging(logger),
		middleware.Metrics(modules.metrics),
		middleware.Secure,
		middleware.CORS(cfg.CORSOrigins, cfg.IsDevelopment()),
	)

	registerHealthRoutes(router, modules)
	router.Handle("/metrics", metrics.Handler(modules.metrics))

	organizationAccess := func(resource string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			requireAuthenticated := modules.organizationsContext.RequireAuthenticated(modules.authn)
			requireAccess := modules.organizationsContext.RequireAccess(resource)
			return requireAuthenticated(
				modules.platform.Networking.Middleware.Enforce(
					modules.rateLimit.Handle(requireAccess(next)),
				),
			)
		}
	}
	sessionOrganizationAccess := func(resource string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			requireAccess := modules.organizationsContext.RequireAccess(resource)
			return modules.authn.RequireSession(
				modules.organizationsContext.Require(
					modules.platform.Networking.Middleware.Enforce(
						modules.rateLimit.Handle(requireAccess(next)),
					),
				),
			)
		}
	}
	organizationContextAccess := func(resource string) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			requireAccess := modules.organizationsContext.RequireAccess(resource)
			return modules.organizationsContext.Require(
				modules.platform.Networking.Middleware.Enforce(
					modules.rateLimit.Handle(requireAccess(next)),
				),
			)
		}
	}

	router.Route("/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(modules.rateLimit.HandleAuth)
			identity.RegisterRoutes(r, modules.identity, modules.authn.RequireSession)
		})
		tenancy.RegisterRoutes(
			r,
			modules.tenancy,
			modules.authn.RequireSession,
			organizationContextAccess,
			sessionOrganizationAccess,
		)
		platform.RegisterRoutes(r, modules.platform, organizationAccess)
		ai.RegisterRoutes(r, modules.ai, organizationAccess)
		telephony.RegisterRoutes(
			r,
			modules.telephony,
			organizationAccess,
			modules.platform.Idempotency.Middleware.Handle,
		)
	})

	return router
}
