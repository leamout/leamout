package edge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestHandlerReturnsForbiddenWhenRouteDoesNotAuthorize(t *testing.T) {
	service := NewService(&fakeStore{resolveErr: pgx.ErrNoRows}, &fakeState{})
	handler := NewHandler(service, "edge-secret")
	request := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/internal/v1/sip-edge/authorize", strings.NewReader(`{
		"username":"managed-trunk",
		"realm":"sip.leamout.com",
		"from":"+14155550100",
		"to":"+14155550101"
	}`))
	request.Header.Set("Authorization", "Bearer edge-secret")
	recorder := httptest.NewRecorder()

	handler.Admit(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
	var response struct {
		Allowed bool `json:"allowed"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Allowed {
		t.Fatal("denied response was marked allowed")
	}
}
