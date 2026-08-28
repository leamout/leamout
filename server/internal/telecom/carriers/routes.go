package carriers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, handler *Handler, auth func(http.Handler) http.Handler) {
	router.Route("/carrier-connections", func(r chi.Router) {
		r.Use(auth)
		r.Post("/", handler.Create)
		r.Get("/", handler.List)
		r.Get("/{carrier_connection_id}", handler.Get)
		r.Patch("/{carrier_connection_id}", handler.Update)
		r.Delete("/{carrier_connection_id}", handler.Delete)
		r.Post("/{carrier_connection_id}/source-ips", handler.CreateSourceIP)
		r.Get("/{carrier_connection_id}/source-ips", handler.ListSourceIPs)
		r.Delete("/{carrier_connection_id}/source-ips/{source_ip_id}", handler.DeleteSourceIP)
	})
}
