package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

type fakeReadinessDatabase struct {
	err error
}

func (f fakeReadinessDatabase) Ping(context.Context) error {
	return f.err
}

type fakeReadinessMedia struct {
	err error
}

func (f fakeReadinessMedia) HealthCheck(context.Context) error {
	return f.err
}

func TestHealthRoutes(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		database   error
		media      error
		wantStatus int
	}{
		{
			name:       "liveness ignores dependency failure",
			path:       "/healthz",
			database:   errors.New("database unavailable"),
			media:      errors.New("freeswitch unavailable"),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "readiness succeeds when dependencies are healthy",
			path:       "/readyz",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "readiness fails when database is unavailable",
			path:       "/readyz",
			database:   errors.New("database unavailable"),
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "readiness fails when freeswitch is unavailable",
			path:       "/readyz",
			media:      errors.New("freeswitch unavailable"),
			wantStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			RegisterHealthRoutes(
				router,
				fakeReadinessDatabase{err: tt.database},
				fakeReadinessMedia{err: tt.media},
			)

			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}
