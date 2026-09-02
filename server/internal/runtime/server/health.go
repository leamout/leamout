package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

const readinessTimeout = 2 * time.Second

type readinessDatabase interface {
	Ping(context.Context) error
}

type readinessMedia interface {
	HealthCheck(context.Context) error
}

type readinessCache interface {
	Ping(context.Context) error
}

type readinessResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func RegisterHealthRoutes(
	router chi.Router,
	database readinessDatabase,
	cache readinessCache,
	media readinessMedia,
) {
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	router.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readinessTimeout)
		defer cancel()

		checks := map[string]string{
			"database":   "ok",
			"redis":      "ok",
			"freeswitch": "ok",
		}
		status := http.StatusNoContent

		if database == nil || database.Ping(ctx) != nil {
			checks["database"] = "unavailable"
			status = http.StatusServiceUnavailable
		}
		if cache == nil || cache.Ping(ctx) != nil {
			checks["redis"] = "unavailable"
			status = http.StatusServiceUnavailable
		}
		if media == nil || media.HealthCheck(ctx) != nil {
			checks["freeswitch"] = "unavailable"
			status = http.StatusServiceUnavailable
		}

		if status == http.StatusNoContent {
			w.WriteHeader(status)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(readinessResponse{
			Status: "unavailable",
			Checks: checks,
		})
	})
}
