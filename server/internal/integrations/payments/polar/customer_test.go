package polar

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

type customerStateTransport func(*http.Request) (*http.Response, error)

type customerStateContextKey struct{}

func (transport customerStateTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func TestGetCustomerStateByExternalID(t *testing.T) {
	externalID := "organization/id+with spaces"
	ctx := context.WithValue(t.Context(), customerStateContextKey{}, "customer-state")
	const response = `{
		"id": "43814acd-2e4e-4480-b8c5-6a14452089a9",
		"external_id": "organization/id+with spaces",
		"active_subscriptions": [
			{
				"id": "34e15c3d-bfa6-4ee4-a0ec-60a7c5e94e30",
				"product_id": "da5b3818-e92e-4b09-afad-ae61895f05db",
				"status": "active",
				"current_period_start": "2026-10-01T00:00:00Z",
				"current_period_end": "2026-11-01T00:00:00Z",
				"trial_end": null,
				"cancel_at_period_end": true,
				"canceled_at": "2026-10-13T10:00:00Z"
			},
			{
				"id": "c4c42e98-8ea1-48db-8222-ce60eac69222",
				"product_id": "58ff657a-74d9-4342-89c4-2a3b784c6965",
				"status": "trialing",
				"current_period_start": "2026-10-13T00:00:00Z",
				"current_period_end": "2026-10-20T00:00:00Z",
				"trial_end": "2026-10-20T00:00:00Z",
				"cancel_at_period_end": false,
				"canceled_at": null
			}
		]
	}`
	httpClient := &http.Client{
		Transport: customerStateTransport(func(request *http.Request) (*http.Response, error) {
			if request.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", request.Method)
			}
			if request.URL.Host != "sandbox-api.polar.sh" {
				t.Errorf("host = %q, want sandbox API", request.URL.Host)
			}
			if request.URL.EscapedPath() != "/v1/customers/external/"+url.PathEscape(externalID)+"/state" {
				t.Errorf("escaped path = %q", request.URL.EscapedPath())
			}
			if request.Header.Get("Authorization") != "Bearer test-token" {
				t.Error("missing Polar authentication")
			}
			if request.Header.Get("Polar-Version") != APIVersion {
				t.Error("missing pinned Polar API version")
			}
			if request.Context().Value(customerStateContextKey{}) != "customer-state" {
				t.Error("customer state request did not preserve context")
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(response)),
			}, nil
		}),
	}
	client := NewWithHTTPClient("test-token", "test-secret", "sandbox", httpClient)
	state, err := client.GetCustomerStateByExternalID(ctx, externalID)
	if err != nil {
		t.Fatal(err)
	}
	if state.ID != "43814acd-2e4e-4480-b8c5-6a14452089a9" || state.ExternalID != externalID {
		t.Fatalf("incorrect customer identity: %+v", state)
	}
	if len(state.ActiveSubscriptions) != 2 {
		t.Fatalf("subscriptions = %d, want 2", len(state.ActiveSubscriptions))
	}
	active := state.ActiveSubscriptions[0]
	if active.ProductID != ProProductID || active.Status != "active" {
		t.Fatalf("incorrect active subscription: %+v", active)
	}
	if !active.CancelAtPeriodEnd || active.CanceledAt == nil || active.TrialEnd != nil {
		t.Fatalf("incorrect cancellation fields: %+v", active)
	}
	if !active.CurrentPeriodEnd.Equal(time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("incorrect current period end: %v", active.CurrentPeriodEnd)
	}
	trialing := state.ActiveSubscriptions[1]
	if trialing.Status != "trialing" || trialing.TrialEnd == nil {
		t.Fatalf("incorrect trialing subscription: %+v", trialing)
	}
}

func TestGetCustomerStateRejectsInvalidIdentity(t *testing.T) {
	tests := []struct {
		name       string
		externalID string
		response   string
		requests   int
	}{
		{
			name:       "empty external ID",
			externalID: " ",
		},
		{
			name:       "missing customer ID",
			externalID: "organization",
			response:   `{"external_id":"organization","active_subscriptions":[]}`,
			requests:   1,
		},
		{
			name:       "incorrect external ID",
			externalID: "organization",
			response:   `{"id":"43814acd-2e4e-4480-b8c5-6a14452089a9","external_id":"another","active_subscriptions":[]}`,
			requests:   1,
		},
		{
			name:       "malformed response",
			externalID: "organization",
			response:   `{`,
			requests:   1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requests := 0
			httpClient := &http.Client{
				Transport: customerStateTransport(func(_ *http.Request) (*http.Response, error) {
					requests++
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(strings.NewReader(test.response)),
					}, nil
				}),
			}
			client := NewWithHTTPClient("test-token", "test-secret", "production", httpClient)
			if _, err := client.GetCustomerStateByExternalID(t.Context(), test.externalID); err == nil {
				t.Fatal("invalid customer state was accepted")
			}
			if requests != test.requests {
				t.Fatalf("requests = %d, want %d", requests, test.requests)
			}
		})
	}
}

func TestCustomerStateAPIErrors(t *testing.T) {
	for _, status := range []int{
		http.StatusNotFound,
		http.StatusUnauthorized,
		http.StatusTooManyRequests,
		http.StatusServiceUnavailable,
	} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			httpClient := &http.Client{
				Transport: customerStateTransport(func(_ *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: status,
						Body:       io.NopCloser(strings.NewReader(`{"detail":"provider-secret"}`)),
					}, nil
				}),
			}
			client := NewWithHTTPClient("access-token-secret", "test-secret", "production", httpClient)
			_, err := client.GetCustomerStateByExternalID(t.Context(), "organization")
			var apiError *APIError
			if !errors.As(err, &apiError) || apiError.StatusCode != status {
				t.Fatalf("error = %v, want API status %d", err, status)
			}
			if errors.Is(err, ErrNotFound) != (status == http.StatusNotFound) {
				t.Fatalf("error = %v, incorrect not-found classification", err)
			}
			if strings.Contains(err.Error(), "secret") {
				t.Fatalf("error exposed sensitive data: %v", err)
			}
		})
	}
}

func TestNewWithHTTPClientRetainsRedirectProtection(t *testing.T) {
	provided := &http.Client{}
	client := NewWithHTTPClient("test-token", "test-secret", "production", provided)
	if provided.Timeout != 0 || provided.CheckRedirect != nil {
		t.Fatal("provided HTTP client was mutated")
	}
	if client.httpClient.Timeout != 15*time.Second {
		t.Fatalf("timeout = %v, want 15s", client.httpClient.Timeout)
	}
	if err := client.httpClient.CheckRedirect(nil, nil); !errors.Is(err, http.ErrUseLastResponse) {
		t.Fatalf("redirect policy = %v, want ErrUseLastResponse", err)
	}
}

func TestPolarClientEnvironment(t *testing.T) {
	for _, environment := range []string{
		"production",
		"sandbox",
	} {
		t.Run(environment, func(t *testing.T) {
			client := NewWithHTTPClient("test-token", "test-secret", environment, nil)
			if client.Environment() != environment {
				t.Fatalf("environment = %q, want %q", client.Environment(), environment)
			}
		})
	}
}

func TestSubscriptionDecodesProviderTimestamps(t *testing.T) {
	var subscription Subscription
	if err := json.Unmarshal([]byte(`{
		"modified_at": "2026-10-13T10:00:00Z",
		"past_due_at": "2026-10-12T10:00:00Z"
	}`), &subscription); err != nil {
		t.Fatal(err)
	}
	if subscription.ModifiedAt == nil || subscription.PastDueAt == nil {
		t.Fatalf("missing provider timestamps: %+v", subscription)
	}
	if !subscription.ModifiedAt.Equal(time.Date(2026, time.October, 13, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("modified at = %v", subscription.ModifiedAt)
	}
	if !subscription.PastDueAt.Equal(time.Date(2026, time.October, 12, 10, 0, 0, 0, time.UTC)) {
		t.Fatalf("past due at = %v", subscription.PastDueAt)
	}
	if err := json.Unmarshal([]byte(`{
		"modified_at": null,
		"past_due_at": null
	}`), &subscription); err != nil {
		t.Fatal(err)
	}
	if subscription.ModifiedAt != nil || subscription.PastDueAt != nil {
		t.Fatalf("null provider timestamps were retained: %+v", subscription)
	}
}
