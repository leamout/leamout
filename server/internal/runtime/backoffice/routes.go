package backoffice

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/internal/backoffice/assets"
	backofficecalls "github.com/leamout/leamout/internal/backoffice/calls"
	"github.com/leamout/leamout/internal/backoffice/dashboard"
	backofficeorganizations "github.com/leamout/leamout/internal/backoffice/organizations"
)

func registerRoutes(router chi.Router) {
	router.Get("/healthz", health)

	public := http.FileServerFS(assets.Public())
	router.Handle("/favicon.ico", public)
	router.Handle("/static/*", public)

	dashboard.NewHandler().Routes(router)
	router.Mount("/organizations", backofficeorganizations.NewHandler().Routes())
	router.Mount("/calls", backofficecalls.NewHandler().Routes())
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}
