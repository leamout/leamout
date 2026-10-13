package polar

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
)

func TestCheckoutAndPortalUsePolarAPIContract(t *testing.T) {
	organizationID := uuid.New().String()
	customerID := uuid.New().String()
	subscriptionID := uuid.New().String()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" || r.Header.Get("Polar-Version") != APIVersion {
			t.Error("missing Polar authentication or pinned API version")
		}
		switch r.URL.Path {
		case "/v1/checkouts/":
			var input struct {
				Products           []string          `json:"products"`
				ExternalCustomerID string            `json:"external_customer_id"`
				ReturnURL          string            `json:"return_url"`
				Metadata           map[string]string `json:"metadata"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				t.Error(err)
			}
			if r.Method != http.MethodPost || len(input.Products) != 1 || input.Products[0] != DeveloperProductID || input.ExternalCustomerID != organizationID || input.Metadata["organization_id"] != organizationID || input.ReturnURL != "https://leamout.com/billing" {
				t.Error("incorrect Polar checkout payload")
			}
			_, _ = w.Write([]byte(`{"id":"checkout","url":"https://polar.sh/checkout/test"}`))
		case "/v1/customer-sessions/":
			var input PortalParams
			// Use the actual JSON field names in the API request.
			var wire struct {
				CustomerID string `json:"customer_id"`
				ReturnURL  string `json:"return_url"`
			}
			if err := json.NewDecoder(r.Body).Decode(&wire); err != nil {
				t.Error(err)
			}
			input.CustomerID = wire.CustomerID
			input.ReturnURL = wire.ReturnURL
			if input.CustomerID != customerID || input.ReturnURL != "https://leamout.com/billing" {
				t.Error("incorrect portal customer")
			}
			_, _ = w.Write([]byte(`{"customer_portal_url":"https://polar.sh/portal/test","token":"must-not-leak"}`))
		case "/v1/subscriptions/" + subscriptionID:
			if r.Method != http.MethodGet {
				t.Error("subscription lookup is not GET")
			}
			_, _ = w.Write([]byte(`{"id":"` + subscriptionID + `","customer_id":"` + customerID + `","product_id":"` + ProProductID + `"}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client := New("test-token", "test-secret", "production")
	client.baseURL = server.URL
	checkout, err := client.CreateCheckoutSession(t.Context(), CheckoutParams{
		ProductID:      DeveloperProductID,
		OrganizationID: organizationID,
		PlanCode:       "developer",
		SuccessURL:     "https://leamout.com/success",
		ReturnURL:      "https://leamout.com/billing",
	})
	if err != nil || checkout.URL != "https://polar.sh/checkout/test" {
		t.Fatalf("checkout=%+v error=%v", checkout, err)
	}
	portal, err := client.CreatePortalSession(t.Context(), PortalParams{
		CustomerID: customerID,
		ReturnURL:  "https://leamout.com/billing",
	})
	if err != nil || portal.URL != "https://polar.sh/portal/test" {
		t.Fatalf("portal=%+v error=%v", portal, err)
	}
	subscription, err := client.GetSubscription(t.Context(), subscriptionID)
	if err != nil || subscription.ProductID != ProProductID {
		t.Fatalf("subscription=%+v error=%v", subscription, err)
	}
}
