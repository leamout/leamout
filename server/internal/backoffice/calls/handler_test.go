package calls

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type stubRepository struct {
	detail Detail
	err    error
}

func (s stubRepository) List(context.Context) ([]Call, error) { return nil, nil }
func (s stubRepository) Get(context.Context, uuid.UUID) (Detail, error) {
	return s.detail, s.err
}

func detailRouter(repository repository) http.Handler {
	handler := NewHandler(repository)
	router := chi.NewRouter()
	router.Mount("/calls", handler.Routes())
	return router
}

func TestDetailRejectsMalformedID(t *testing.T) {
	recorder := httptest.NewRecorder()
	detailRouter(stubRepository{}).ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/calls/not-a-uuid", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestDetailReturnsNotFoundForUnknownCall(t *testing.T) {
	recorder := httptest.NewRecorder()
	detailRouter(stubRepository{err: pgx.ErrNoRows}).ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/calls/"+uuid.NewString(), nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestDetailRendersSafeOperatorProjection(t *testing.T) {
	callID := uuid.NewString()
	recorder := httptest.NewRecorder()
	detailRouter(stubRepository{detail: Detail{Call: Call{ID: callID, OrganizationID: uuid.NewString(), Organization: "Acme", From: "sip:alice@example.test", To: "+15551234567", Direction: "outbound", Duration: "00:01:02", Status: "completed", CreatedAt: "2026-09-10"}, MediaState: "active", RecordingStatus: "completed"}}).ServeHTTP(recorder, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/calls/"+callID, nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	for _, want := range []string{"Call lifecycle", "Routing", "Acme", "sip:alice@example.test"} {
		if !strings.Contains(recorder.Body.String(), want) {
			t.Errorf("response missing %q", want)
		}
	}
}
