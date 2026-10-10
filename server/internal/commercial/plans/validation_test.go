package plans

import "testing"

func TestValidateUpsert(t *testing.T) {
	t.Parallel()

	req, err := validateUpsert(UpsertRequest{
		Code:            " pro ",
		Name:            " Pro ",
		Currency:        " usd ",
		AmountMinor:     4900,
		BillingInterval: IntervalMonth,
		Entitlements: map[string]bool{
			"advanced_rbac":      true,
			"retention_policies": true,
		},
		Status: StatusActive,
	})
	if err != nil {
		t.Fatalf("validateUpsert() error = %v", err)
	}
	if req.Code != "pro" || req.Name != "Pro" || req.Currency != "USD" {
		t.Fatalf("validateUpsert() normalization = %#v", req)
	}
}

func TestValidateUpsertRejectsInvalidAmount(t *testing.T) {
	t.Parallel()

	_, err := validateUpsert(UpsertRequest{
		Code:            "pro",
		Name:            "Pro",
		Currency:        "USD",
		AmountMinor:     -1,
		BillingInterval: IntervalMonth,
		Status:          StatusActive,
	})
	if err == nil {
		t.Fatal("expected negative plan amount to be rejected")
	}
}

func TestValidateUpsertDefaultsEntitlements(t *testing.T) {
	t.Parallel()

	req, err := validateUpsert(UpsertRequest{
		Code:            "builder",
		Name:            "Builder",
		Currency:        "USD",
		AmountMinor:     4900,
		BillingInterval: IntervalMonth,
		Status:          StatusActive,
	})
	if err != nil {
		t.Fatalf("validateUpsert() error = %v", err)
	}
	if req.Entitlements == nil {
		t.Fatal("expected entitlements to default to an empty map")
	}
	if len(req.Entitlements) != 0 {
		t.Fatalf("expected no default entitlements, got %#v", req.Entitlements)
	}
}

func TestValidateUpsertRejectsUnknownEntitlement(t *testing.T) {
	t.Parallel()

	_, err := validateUpsert(UpsertRequest{
		Code:            "enterprise",
		Name:            "Enterprise",
		Currency:        "USD",
		AmountMinor:     0,
		BillingInterval: IntervalYear,
		Entitlements: map[string]bool{
			"unknown_capability": true,
		},
		Status: StatusActive,
	})
	if err == nil {
		t.Fatal("expected unknown plan entitlement to be rejected")
	}
}
