package backoffice

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/internal/backoffice/assets"
)

func registerRoutes(router chi.Router, modules Modules) {
	router.Get("/healthz", health)

	public := http.FileServerFS(assets.Public())
	router.Handle("/favicon.ico", public)
	router.Handle("/static/*", public)

	modules.Dashboard.Routes(router)
	router.Mount("/organizations", modules.Organizations.Routes())
	router.Mount("/calls", modules.Calls.Routes())
	router.Mount("/numbers", modules.Numbers.Routes())
	router.Mount("/trunks", modules.Trunks.Routes())
	router.Mount("/carrier-connections", modules.CarrierConnections.Routes())
	router.Mount("/providers", modules.Providers.Routes())
	router.Mount("/commercial", modules.Commercial.Routes())
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}
