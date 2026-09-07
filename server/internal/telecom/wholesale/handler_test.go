package wholesale

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestHandlerReconcilesAuthenticatedCDR(t *testing.T) {
	store := &fakeStore{result: Result{ProviderCDRID: uuid.New(), CallID: uuid.New(), ChargeID: uuid.New()}}
	handler := NewHandler(NewService(store), "edge-secret")
	body, err := json.Marshal(validCDR())
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodPost, "/internal/v1/provider-cdrs/reconcile", bytes.NewReader(body))
	request.Header.Set("Authorization", "Bearer edge-secret")
	recorder := httptest.NewRecorder()

	handler.Reconcile(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestHandlerRejectsUnauthenticatedCDR(t *testing.T) {
	handler := NewHandler(NewService(&fakeStore{}), "edge-secret")
	request := httptest.NewRequest(http.MethodPost, "/internal/v1/provider-cdrs/reconcile", bytes.NewReader([]byte(`{}`)))
	recorder := httptest.NewRecorder()

	handler.Reconcile(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}
