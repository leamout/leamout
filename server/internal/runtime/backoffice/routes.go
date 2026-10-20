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

func newRouter(modules *modules, logger *logging.Logger, frontendURL string) (*chi.Mux, error) {
	// Trust the configured public console origin, never forwarded request headers.
	frontend, err := url.Parse(frontendURL)
	if err != nil || frontend.Host == "" || frontend.User != nil || (frontend.Scheme != "http" && frontend.Scheme != "https") {
		return nil, fmt.Errorf("frontend URL must be an absolute HTTP(S) URL")
	}
	protection := http.NewCrossOriginProtection()
	if err := protection.AddTrustedOrigin(frontend.Scheme + "://" + frontend.Host); err != nil {
		return nil, fmt.Errorf("configure backoffice origin protection: %w", err)
	}
	router := chi.NewRouter()
	router.Use(chimiddleware.Recoverer)
	router.Use(securityHeaders)
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})
	static := http.StripPrefix("/backoffice", http.FileServerFS(assets.Public()))
	router.Handle("/backoffice/static/*", static)
	router.Handle("/backoffice/favicon.ico", static)
	router.Route("/backoffice", func(protected chi.Router) {
		protected.Use(modules.authn.RequireSession)
		protected.Use(requirePlatformAdmin(modules.queries, logger))
		protected.Use(protection.Handler)
		protected.Get("/", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/backoffice/organizations", http.StatusSeeOther)
		})
		modules.organizations.Routes(protected)
	})
	return router, nil
}
