package backoffice

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/internal/backoffice/assets"
)

type accessMiddleware func(http.Handler) http.Handler

func registerRoutes(
	router chi.Router,
	modules Modules,
	access ...accessMiddleware,
) {
	router.Get("/healthz", health)

	public := http.FileServerFS(assets.Public())
	router.Handle("/favicon.ico", public)
	router.Handle("/static/*", public)

	router.Group(func(protected chi.Router) {
		for _, middleware := range access {
			if middleware != nil {
				protected.Use(middleware)
			}
		}

		modules.Dashboard.Routes(protected)
		protected.Mount("/users", modules.Users.Routes())
		protected.Mount("/organizations", modules.Organizations.Routes())
		protected.Mount("/calls", modules.Calls.Routes())
		protected.Mount("/numbers", modules.Numbers.Routes())
		protected.Mount("/trunks", modules.Trunks.Routes())
		protected.Mount("/carrier-connections", modules.CarrierConnections.Routes())
		protected.Mount("/providers", modules.Providers.Routes())
		protected.Mount("/commercial", modules.Commercial.Routes())
		protected.Mount("/subscribers", modules.Subscribers.Routes())
		protected.Mount("/sip-domains", modules.SIPDomains.Routes())
		protected.Mount("/voice-applications", modules.VoiceApplications.Routes())
		protected.Mount("/conferences", modules.Conferences.Routes())
		protected.Mount("/recordings", modules.Recordings.Routes())
		protected.Mount("/provider-cdrs", modules.ProviderCDRs.Routes())
	})
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}
