package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

const healthCheckTimeout = 2 * time.Second

type dependencyChecker interface {
	Ping(context.Context) error
}

type mediaChecker interface {
	HealthCheck(context.Context) error
}

type componentState struct {
	Running   bool      `json:"running"`
	LastError string    `json:"last_error,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type healthState struct {
	mu         sync.RWMutex
	components map[string]componentState
}

func newHealthState(componentNames ...string) *healthState {
	now := time.Now().UTC()
	components := make(map[string]componentState, len(componentNames))
	for _, name := range componentNames {
		components[name] = componentState{UpdatedAt: now}
	}
	return &healthState{components: components}
}

func (s *healthState) setRunning(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.components[name] = componentState{Running: true, UpdatedAt: time.Now().UTC()}
}

func (s *healthState) setStopped(name string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state := componentState{UpdatedAt: time.Now().UTC()}
	if err != nil {
		state.LastError = err.Error()
	}
	s.components[name] = state
}

func (s *healthState) snapshot() map[string]componentState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]componentState, len(s.components))
	for name, state := range s.components {
		result[name] = state
	}
	return result
}

func healthHandler(database, nats, cache dependencyChecker, media mediaChecker, state *healthState) http.Handler {
	router := http.NewServeMux()
	router.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	router.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), healthCheckTimeout)
		defer cancel()
		checks := map[string]string{}
		ready := true
		for name, checker := range map[string]dependencyChecker{"database": database, "nats": nats, "redis": cache} {
			if checker == nil || checker.Ping(ctx) != nil {
				checks[name] = "unavailable"
				ready = false
			} else {
				checks[name] = "ok"
			}
		}
		if media == nil || media.HealthCheck(ctx) != nil {
			checks["freeswitch"] = "unavailable"
			ready = false
		} else {
			checks["freeswitch"] = "ok"
		}
		components := state.snapshot()
		for _, component := range components {
			if !component.Running {
				ready = false
			}
		}
		if ready {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "unavailable", "checks": checks, "components": components})
	})
	router.HandleFunc("GET /metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		components := state.snapshot()
		names := make([]string, 0, len(components))
		for name := range components {
			names = append(names, name)
		}
		sort.Strings(names)
		_, _ = fmt.Fprintln(w, "# HELP leamout_worker_component_up Whether a worker component is running.")
		_, _ = fmt.Fprintln(w, "# TYPE leamout_worker_component_up gauge")
		for _, name := range names {
			value := 0
			if components[name].Running {
				value = 1
			}
			_, _ = fmt.Fprintf(w, "leamout_worker_component_up{component=%q} %d\n", name, value)
		}
	})
	return router
}
