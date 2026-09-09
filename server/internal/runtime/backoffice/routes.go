package backoffice

import (
	"bytes"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	backofficeui "github.com/leamout/leamout/internal/backoffice"
	backofficecalls "github.com/leamout/leamout/internal/backoffice/calls"
	"github.com/leamout/leamout/internal/backoffice/components"
	backofficeorganizations "github.com/leamout/leamout/internal/backoffice/organizations"
)

func registerRoutes(router chi.Router) {
	router.Get("/healthz", health)
	public := http.FileServerFS(backofficeui.PublicAssets())
	router.Get("/", component(components.Dashboard()))
	router.Get("/fragments/runtime-status", component(components.RuntimeStatus()))
	router.Mount("/organizations", backofficeorganizations.NewHandler().Routes())
	router.Mount("/calls", backofficecalls.NewHandler().Routes())
	router.Handle("/favicon.ico", public)
	router.Handle("/static/*", public)
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func component(view templ.Component) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body bytes.Buffer
		if err := view.Render(r.Context(), &body); err != nil {
			http.Error(w, "render backoffice view", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(body.Bytes())
	}
}
