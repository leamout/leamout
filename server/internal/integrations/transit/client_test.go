package transit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestSearchAvailableUsesDeploymentAuthenticationAndOpaqueSelections(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != availableNumbersPath {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer deployment-token" {
			t.Fatalf("authorization = %q", got)
		}
		if got := r.Header.Get("X-Leamout-Deployment-ID"); got != "deployment-123" {
			t.Fatalf("deployment id = %q", got)
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request["country_code"] != "US" || request["contains"] != "212" {
			t.Fatalf("request = %#v", request)
		}
		for _, forbidden := range []string{"provider", "provider_inventory_id", "available_did_id", "provider_product_id", "sku_id"} {
			if _, ok := request[forbidden]; ok {
				t.Fatalf("request leaked %q: %#v", forbidden, request)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"numbers":[{"selection_id":"msel_abc","number":"+12125550123","country_code":"US","voice_enabled":true}]}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{
		BaseURL:      server.URL,
		Token:        "deployment-token",
		DeploymentID: "deployment-123",
	})
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.SearchAvailable(context.Background(), AvailableNumberSearchRequest{
		CountryCode: "US",
		Contains:    "212",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(response.Numbers) != 1 || response.Numbers[0].SelectionID != "msel_abc" {
		t.Fatalf("response = %#v", response)
	}
}

func TestExecuteNumberOrderSendsOnlyOperationAndSelectionHandles(t *testing.T) {
	operationID := uuid.New()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != numberOrdersPath {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		var request map[string]any
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatal(err)
		}
		if request["operation_id"] != operationID.String() || request["selection_id"] != "msel_abc" {
			t.Fatalf("request = %#v", request)
		}
		if len(request) != 2 {
			t.Fatalf("request contains unexpected provider metadata: %#v", request)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"state":"completed","managed_resource_id":"mnum_123","number":"+12125550123","country_code":"US"}`))
	}))
	defer server.Close()

	client, err := NewClient(Config{
		BaseURL:      server.URL,
		Token:        "deployment-token",
		DeploymentID: "deployment-123",
	})
	if err != nil {
		t.Fatal(err)
	}
	response, err := client.ExecuteNumberOrder(context.Background(), ExecuteNumberOrderRequest{
		OperationID: operationID,
		SelectionID: "msel_abc",
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.State != "completed" || response.ManagedResourceID != "mnum_123" {
		t.Fatalf("response = %#v", response)
	}
}

func TestNewClientRequiresDeploymentCredentials(t *testing.T) {
	if _, err := NewClient(Config{BaseURL: "https://transit.example.test", DeploymentID: "deployment-123"}); err == nil {
		t.Fatal("expected missing token error")
	}
	if _, err := NewClient(Config{BaseURL: "https://transit.example.test", Token: "token"}); err == nil {
		t.Fatal("expected missing deployment id error")
	}
}
