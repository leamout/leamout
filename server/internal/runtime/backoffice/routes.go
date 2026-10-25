package backoffice

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/leamout/leamout/server/internal/backoffice/assets"
	"github.com/leamout/leamout/server/internal/observability/logging"
)

func newRouter(modules *modules, logger *logging.Logger, backofficeURL string) (*chi.Mux, error) {
	// Trust the configured public backoffice origin, never forwarded request headers.
	origin, err := url.Parse(backofficeURL)
	if err != nil || origin.Host == "" || origin.User != nil || (origin.Scheme != "http" && origin.Scheme != "https") {
		return nil, fmt.Errorf("backoffice URL must be an absolute HTTP(S) URL")
	}
	protection := http.NewCrossOriginProtection()
	if err := protection.AddTrustedOrigin(origin.Scheme + "://" + origin.Host); err != nil {
		return nil, fmt.Errorf("configure backoffice origin protection: %w", err)
	}
	router := chi.NewRouter()
	router.Use(chimiddleware.Recoverer)
	router.Use(securityHeaders)
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	static := http.FileServerFS(assets.Public())
	router.Handle("/static/*", static)
	router.Handle("/favicon.ico", static)
	modules.authentication.Routes(router, protection.Handler)
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/log-in", http.StatusSeeOther)
	})
	router.Group(func(protected chi.Router) {
		protected.Use(modules.authn.RequireSession)
		protected.Use(requirePlatformAdmin(modules.queries, logger))
		protected.Use(protection.Handler)
		protected.Get("/auth/account", modules.authentication.Account)
		modules.organizations.Routes(protected)
		modules.agents.Routes(protected)
		modules.users.Routes(protected)
		modules.numbers.Routes(protected)
		modules.calls.Routes(protected)
		modules.subscriptions.Routes(protected)
	})
	return router, nil
}
