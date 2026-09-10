package calls

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type stubRepository struct {
	calls   []Call
	listErr error
	detail  Detail
	getErr  error
	gotID   uuid.UUID
}

func (r *stubRepository) List(context.Context) ([]Call, error) {
	return r.calls, r.listErr
}

func (r *stubRepository) Get(_ context.Context, id uuid.UUID) (Detail, error) {
	r.gotID = id
	return r.detail, r.getErr
}

func TestDetailMalformedCallIDReturnsNotFound(t *testing.T) {
	handler := NewHandler(&stubRepository{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/not-a-uuid", nil)

	handler.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func TestDetailMissingCallReturnsNotFound(t *testing.T) {
	callID := uuid.New()
	repository := &stubRepository{getErr: pgx.ErrNoRows}
	handler := NewHandler(repository)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/"+callID.String(), nil)

	handler.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
	if repository.gotID != callID {
		t.Fatalf("repository call ID = %s, want %s", repository.gotID, callID)
	}
}

func TestDetailRendersCallProjection(t *testing.T) {
	callID := uuid.New()
	organizationID := uuid.New()
	repository := &stubRepository{detail: Detail{
		Call: Call{
			ID:             callID.String(),
			OrganizationID: organizationID.String(),
			Organization:   "Acme Telecom",
			From:           "+233200000001",
			To:             "+233200000002",
			Direction:      "outbound",
			Duration:       "00:01:30",
			Status:         "completed",
			CreatedAt:      "2026-09-10 09:00:00",
		},
		MediaState:      "active",
		HangupReason:    "normal_clearing",
		RecordingCount:  1,
		RecordingStatus: "completed",
		StartedAt:       "2026-09-10 09:00:01",
		AnsweredAt:      "2026-09-10 09:00:05",
		EndedAt:         "2026-09-10 09:01:35",
		UpdatedAt:       "2026-09-10 09:01:35",
	}}
	handler := NewHandler(repository)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/"+callID.String(), nil)

	handler.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		callID.String(),
		"Acme Telecom",
		"+233200000001",
		"+233200000002",
		"completed",
		"normal_clearing",
		"/organizations/" + organizationID.String(),
	} {
		if !strings.Contains(body, want) {
			t.Errorf("response body does not contain %q", want)
		}
	}
}

func TestIndexLinksCallsAndOrganizations(t *testing.T) {
	callID := uuid.New().String()
	organizationID := uuid.New().String()
	repository := &stubRepository{calls: []Call{{
		ID:             callID,
		OrganizationID: organizationID,
		Organization:   "Acme Telecom",
		From:           "+233200000001",
		To:             "+233200000002",
		Direction:      "outbound",
		Duration:       "00:00:42",
		Status:         "completed",
		CreatedAt:      "2026-09-10 09:00",
	}}}
	handler := NewHandler(repository)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	handler.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		"/calls/" + callID,
		"/organizations/" + organizationID,
		"Acme Telecom",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("response body does not contain %q", want)
		}
	}
}
