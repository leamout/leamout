package billing

import (
	"testing"
	"time"

	"github.com/google/uuid"

	polarintegration "github.com/leamout/leamout/server/internal/integrations/payments/polar"
)

func TestPlanCodeFromPolarProduct(t *testing.T) {
	t.Parallel()
	service := &Service{
		productIDs: map[string]string{
			"developer": polarintegration.DeveloperProductID,
			"pro":       polarintegration.ProProductID,
			"scale":     polarintegration.ScaleProductID,
		},
	}
	for _, tc := range []struct {
		name      string
		productID string
		want      string
		valid     bool
	}{
		{
			"developer",
			polarintegration.DeveloperProductID,
			"developer",
			true,
		},
		{
			"pro",
			polarintegration.ProProductID,
			"pro",
			true,
		},
		{
			"scale",
			polarintegration.ScaleProductID,
			"scale",
			true,
		},
		{
			"unknown",
			"other",
			"",
			false,
		},
		{
			"missing",
			"",
			"",
			false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			subscription := polarintegration.Subscription{
				ProductID: tc.productID,
				Metadata: map[string]any{
					"plan_code": "scale",
				},
			}
			got, err := service.planCodeFromProduct(subscription)
			if (err == nil) != tc.valid || got != tc.want {
				t.Fatalf("got %q, error %v; want %q, valid %v", got, err, tc.want, tc.valid)
			}
		})
	}
}

func TestBillingRedirectOrigin(t *testing.T) {
	t.Parallel()
	service := &Service{
		frontendURL: "https://leamout.com",
	}
	for _, tc := range []struct {
		url   string
		valid bool
	}{
		{
			"https://leamout.com/billing/success",
			true,
		},
		{
			"https://LEAMOUT.com/billing/cancel?next=ok",
			true,
		},
		{
			"https://evil.example/capture",
			false,
		},
		{
			"http://leamout.com/billing",
			false,
		},
		{
			"https://leamout.com.evil.example",
			false,
		},
		{
			"https://leamout.com@evil.example",
			false,
		},
		{
			"//leamout.com",
			false,
		},
		{
			"/billing",
			false,
		},
	} {
		t.Run(tc.url, func(t *testing.T) {
			err := validateRedirects(service.frontendURL, tc.url)
			if (err == nil) != tc.valid {
				t.Fatalf("URL %q: err=%v, valid=%v", tc.url, err, tc.valid)
			}
		})
	}
}

func TestPolarSubscriptionStatusMapping(t *testing.T) {
	for _, tc := range []struct {
		providerStatus string
		expected       string
	}{
		{
			"active",
			"active",
		},
		{
			"trialing",
			"trialing",
		},
		{
			"canceled",
			"canceled",
		},
		{
			"paused",
			"incomplete",
		},
		{
			"past_due",
			"past_due",
		},
		{
			"unknown",
			"incomplete",
		},
	} {
		if got := polarStatus(tc.providerStatus); got != tc.expected {
			t.Fatalf("status %q resolved to %q", tc.providerStatus, got)
		}
	}
}

func TestPolarSyncRejectsConflictingCustomerOwnershipBeforeDatabaseAccess(t *testing.T) {
	service := &Service{}
	organizationID := uuid.New().String()
	customerID := uuid.New().String()
	for _, value := range []polarintegration.Subscription{
		{
			CustomerID: customerID,
			Customer: polarintegration.Customer{
				ID:         customerID,
				ExternalID: "invalid-organization",
			},
		},
		{
			CustomerID: customerID,
			Customer: polarintegration.Customer{
				ID:         customerID,
				ExternalID: organizationID,
			},
			Metadata: map[string]any{
				"organization_id": uuid.New().String(),
			},
		},
		{
			CustomerID: customerID,
			Customer: polarintegration.Customer{
				ID:         uuid.New().String(),
				ExternalID: organizationID,
			},
		},
	} {
		if err := service.syncPolarSubscription(t.Context(), time.Now().UTC(), value); err == nil {
			t.Fatal("conflicting Polar customer ownership accepted")
		}
	}
}
