package plans

import "testing"

func TestValidateUpsert(t *testing.T) {
	t.Parallel()

	amount := int64(4900)
	req, err := validateUpsert(UpsertRequest{
		Code:            " pro ",
		Name:            " Pro ",
		PricingType:     PricingTypeFixed,
		Currency:        " usd ",
		AmountMinor:     &amount,
		BillingInterval: IntervalMonth,
		Entitlements: map[string]bool{
			"advanced_rbac":      true,
			"retention_policies": true,
		},
		Limits: map[string]int64{
			LimitMaxConcurrentCalls: 10,
			LimitRetentionDays:      14,
		},
		Status: StatusActive,
	})
	if err != nil {
		t.Fatalf("validateUpsert() error = %v", err)
	}
	if req.Code != "pro" || req.Name != "Pro" || req.Currency != "USD" {
		t.Fatalf("validateUpsert() normalization = %#v", req)
	}
	if !req.Entitlements["advanced_rbac"] {
		t.Fatal("expected advanced_rbac entitlement to remain true")
	}
}

func TestValidateUpsertRejectsInvalidAmount(t *testing.T) {
	t.Parallel()

	amount := int64(-1)
	_, err := validateUpsert(UpsertRequest{
		Code:            "pro",
		Name:            "Pro",
		PricingType:     PricingTypeFixed,
		Currency:        "USD",
		AmountMinor:     &amount,
		BillingInterval: IntervalMonth,
		Status:          StatusActive,
	})
	if err == nil {
		t.Fatal("expected negative plan amount to be rejected")
	}
}

func TestValidateUpsertDefaultsEntitlementsAndLimits(t *testing.T) {
	t.Parallel()

	amount := int64(4900)
	req, err := validateUpsert(UpsertRequest{
		Code:            "builder",
		Name:            "Builder",
		PricingType:     PricingTypeFixed,
		Currency:        "USD",
		AmountMinor:     &amount,
		BillingInterval: IntervalMonth,
		Status:          StatusActive,
	})
	if err != nil {
		t.Fatalf("validateUpsert() error = %v", err)
	}
	if req.Entitlements == nil || len(req.Entitlements) != 0 {
		t.Fatalf("expected empty entitlements, got %#v", req.Entitlements)
	}
	if req.Limits == nil || len(req.Limits) != 0 {
		t.Fatalf("expected empty limits, got %#v", req.Limits)
	}
}

func TestValidateUpsertRejectsUnknownEntitlement(t *testing.T) {
	t.Parallel()

	amount := int64(19900)
	_, err := validateUpsert(UpsertRequest{
		Code:            "pro",
		Name:            "Pro",
		PricingType:     PricingTypeFixed,
		Currency:        "USD",
		AmountMinor:     &amount,
		BillingInterval: IntervalMonth,
		Entitlements: map[string]bool{
			"unknown_capability": true,
		},
		Status: StatusActive,
	})
	if err == nil {
		t.Fatal("expected unknown plan entitlement to be rejected")
	}
}

func TestValidateUpsertAcceptsCustomPricingWithoutAmount(t *testing.T) {
	t.Parallel()

	_, err := validateUpsert(UpsertRequest{
		Code:            "enterprise",
		Name:            "Enterprise",
		PricingType:     PricingTypeCustom,
		Currency:        "USD",
		BillingInterval: IntervalYear,
		Status:          StatusActive,
	})
	if err != nil {
		t.Fatalf("validateUpsert() error = %v", err)
	}
}

func TestValidateUpsertRejectsUnknownLimit(t *testing.T) {
	t.Parallel()

	amount := int64(4900)
	_, err := validateUpsert(UpsertRequest{
		Code:            "developer",
		Name:            "Developer",
		PricingType:     PricingTypeFixed,
		Currency:        "USD",
		AmountMinor:     &amount,
		BillingInterval: IntervalMonth,
		Limits: map[string]int64{
			"max_agents": 10,
		},
		Status: StatusActive,
	})
	if err == nil {
		t.Fatal("expected unknown plan limit to be rejected")
	}
}
