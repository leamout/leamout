package backoffice

import (
	"bytes"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	backofficeui "github.com/leamout/leamout/internal/backoffice"
)

func registerRoutes(router chi.Router) {
	router.Get("/healthz", health)
	router.Get("/", component(backofficeui.Dashboard()))
	router.Get("/fragments/runtime-status", component(backofficeui.RuntimeStatus()))
	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("internal/backoffice/static"))))
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
