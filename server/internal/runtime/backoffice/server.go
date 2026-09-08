package backoffice

import "github.com/go-chi/chi/v5"

type Server struct {
	Router *chi.Mux
}

func New() *Server {
	router := chi.NewRouter()
	router.Use(securityHeaders)
	registerRoutes(router)
	return &Server{Router: router}
}
