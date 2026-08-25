package server

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func New() (*Server, error) {
	// Build application dependencies.
	// Build modules.
	// Build router.
	// Build HTTP server.

	return &Server{
		httpServer: &http.Server{
			Addr: ":8080",
			// Handler: router,

			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       60 * time.Second,
			MaxHeaderBytes:    1 << 20,
		},
	}, nil
}

func (s *Server) Start() error {
	err := s.httpServer.ListenAndServe()

	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
