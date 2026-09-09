package backoffice

import "github.com/go-chi/chi/v5"

type Server struct {
	Router  *chi.Mux
	Modules Modules
}

func New() *Server {
	modules := newModules()
	router := chi.NewRouter()
	router.Use(securityHeaders)
	registerRoutes(router, modules)
	return &Server{
		Router:  router,
		Modules: modules,
	}
}
