package transit

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/leamout/leamout/internal/database/sqlc"
)

type numberOrderExecutionRepositoryStub struct {
	accepted             int
	processing           int
	recordedFailures     int
	failed               int
	completed            int
	completedSelectionID string
	completedNumber      string
	completedCountryCode string
	completedResourceID  string
}

func (r *numberOrderExecutionRepositoryStub) MarkProviderOperationAccepted(
	context.Context,
	sqlc.ProviderOperation,
	string,
	[]byte,
) error {
	r.accepted++
	return nil
}

func (r *numberOrderExecutionRepositoryStub) MarkNumberOrderProcessing(
	context.Context,
	sqlc.ProviderOperation,
) error {
	r.processing++
	return nil
}

func (r *numberOrderExecutionRepositoryStub) RecordProviderOperationFailure(
	context.Context,
	uuid.UUID,
	error,
) error {
	r.recordedFailures++
	return nil
}

func (r *numberOrderExecutionRepositoryStub) FailProviderOperation(
	context.Context,
	sqlc.ProviderOperation,
	error,
) error {
	r.failed++
	return nil
}

func (r *numberOrderExecutionRepositoryStub) CompleteTransitProviderOperation(
	_ context.Context,
	_ sqlc.ProviderOperation,
	selectionID string,
	number string,
	countryCode string,
	managedResourceID string,
	_ []byte,
) error {
	r.completed++
	r.completedSelectionID = selectionID
	r.completedNumber = number
	r.completedCountryCode = countryCode
	r.completedResourceID = managedResourceID
	return nil
}

func TestNumberOrderExecutorCompletesTransitOperation(t *testing.T) {
	operationID := uuid.New()
	numberOrderID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request["operation_id"] != operationID.String() || request["selection_id"] != "msel_abc" {
			t.Fatalf("request = %#v", request)
		}
		if len(request) != 2 {
			t.Fatalf("transit request leaked provider metadata: %#v", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"state":"completed","managed_resource_id":"mnum_123","number":"+12125550123","country_code":"US"}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{
		BaseURL:      server.URL,
		Token:        "token",
		DeploymentID: "deployment-123",
	})
	if err != nil {
		t.Fatal(err)
	}
	repo := &numberOrderExecutionRepositoryStub{}
	executor, err := NewNumberOrderExecutor(client, repo)
	if err != nil {
		t.Fatal(err)
	}

	request, err := json.Marshal(numberOrderOperationRequest{
		SelectionID: "msel_abc",
		Number:      "+12125550123",
		CountryCode: "US",
	})
	if err != nil {
		t.Fatal(err)
	}
	operation := sqlc.ProviderOperation{
		ID:              operationID,
		OrganizationID:  uuid.New(),
		ExecutionTarget: "transit",
		OperationType:   "number_order",
		NumberOrderID:   &numberOrderID,
		State:           "pending",
		Request:         request,
	}
	if err := executor.ExecuteProviderOperation(context.Background(), operation); err != nil {
		t.Fatal(err)
	}

	if repo.accepted != 1 || repo.processing != 1 || repo.completed != 1 {
		t.Fatalf("repo calls accepted=%d processing=%d completed=%d", repo.accepted, repo.processing, repo.completed)
	}
	if repo.failed != 0 || repo.recordedFailures != 0 {
		t.Fatalf("unexpected failure calls failed=%d retries=%d", repo.failed, repo.recordedFailures)
	}
	if repo.completedSelectionID != "msel_abc" ||
		repo.completedNumber != "+12125550123" ||
		repo.completedCountryCode != "US" ||
		repo.completedResourceID != "mnum_123" {
		t.Fatalf("completion = %#v", repo)
	}
}

func TestNumberOrderExecutorRetriesTransientTransitFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "temporarily unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client, err := NewClient(Config{
		BaseURL:      server.URL,
		Token:        "token",
		DeploymentID: "deployment-123",
	})
	if err != nil {
		t.Fatal(err)
	}
	repo := &numberOrderExecutionRepositoryStub{}
	executor, err := NewNumberOrderExecutor(client, repo)
	if err != nil {
		t.Fatal(err)
	}

	numberOrderID := uuid.New()
	request, _ := json.Marshal(numberOrderOperationRequest{
		SelectionID: "msel_abc",
		Number:      "+12125550123",
		CountryCode: "US",
	})
	err = executor.ExecuteProviderOperation(context.Background(), sqlc.ProviderOperation{
		ID:              uuid.New(),
		OrganizationID:  uuid.New(),
		ExecutionTarget: "transit",
		OperationType:   "number_order",
		NumberOrderID:   &numberOrderID,
		State:           "pending",
		Request:         request,
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.recordedFailures != 1 || repo.failed != 0 {
		t.Fatalf("failure calls retries=%d failed=%d", repo.recordedFailures, repo.failed)
	}
}

func TestNumberOrderExecutorFailsNonRetryableTransitFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "invalid selection", http.StatusBadRequest)
	}))
	defer server.Close()

	client, err := NewClient(Config{
		BaseURL:      server.URL,
		Token:        "token",
		DeploymentID: "deployment-123",
	})
	if err != nil {
		t.Fatal(err)
	}
	repo := &numberOrderExecutionRepositoryStub{}
	executor, err := NewNumberOrderExecutor(client, repo)
	if err != nil {
		t.Fatal(err)
	}

	numberOrderID := uuid.New()
	request, _ := json.Marshal(numberOrderOperationRequest{
		SelectionID: "msel_abc",
		Number:      "+12125550123",
		CountryCode: "US",
	})
	err = executor.ExecuteProviderOperation(context.Background(), sqlc.ProviderOperation{
		ID:              uuid.New(),
		OrganizationID:  uuid.New(),
		ExecutionTarget: "transit",
		OperationType:   "number_order",
		NumberOrderID:   &numberOrderID,
		State:           "pending",
		Request:         request,
	})
	if err != nil {
		t.Fatal(err)
	}
	if repo.failed != 1 || repo.recordedFailures != 0 {
		t.Fatalf("failure calls failed=%d retries=%d", repo.failed, repo.recordedFailures)
	}
}

func TestHTTPErrorRetryability(t *testing.T) {
	cases := []struct {
		status    int
		retryable bool
	}{
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
		{http.StatusRequestTimeout, true},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusServiceUnavailable, true},
	}
	for _, tc := range cases {
		err := &HTTPError{StatusCode: tc.status, Message: errors.New("failure").Error()}
		if got := err.Retryable(); got != tc.retryable {
			t.Fatalf("status %d retryable=%v want %v", tc.status, got, tc.retryable)
		}
	}
}
