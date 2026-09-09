package backoffice

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/internal/backoffice/assets"
	backofficeauth "github.com/leamout/leamout/internal/backoffice/auth"
)

func registerRoutes(
	router chi.Router,
	modules Modules,
	authHandler *backofficeauth.Handler,
	authentication authenticator,
) {
	router.Get("/healthz", health)

	public := http.FileServerFS(assets.Public())
	router.Handle("/favicon.ico", public)
	router.Handle("/static/*", public)

	if authHandler != nil {
		authHandler.RegisterPublicRoutes(router)
	}

	router.Group(func(protected chi.Router) {
		protected.Use(requireBackoffice(authentication))

		if authHandler != nil {
			authHandler.RegisterProtectedRoutes(protected)
		}

		modules.Dashboard.Routes(protected)
		protected.Mount("/organizations", modules.Organizations.Routes())
		protected.Mount("/calls", modules.Calls.Routes())
		protected.Mount("/numbers", modules.Numbers.Routes())
		protected.Mount("/trunks", modules.Trunks.Routes())
		protected.Mount("/carrier-connections", modules.CarrierConnections.Routes())
		protected.Mount("/providers", modules.Providers.Routes())
		protected.Mount("/commercial", modules.Commercial.Routes())
	})
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}
