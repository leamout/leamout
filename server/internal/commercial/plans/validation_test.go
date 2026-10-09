package plans

import "testing"

func TestValidateUpsert(t *testing.T) {
	req, err := validateUpsert(UpsertRequest{
		Code:            " pro ",
		Name:            " Pro ",
		Currency:        " usd ",
		AmountMinor:     4900,
		BillingInterval: IntervalMonth,
		Status:          StatusActive,
	})
	if err != nil {
		t.Fatalf("validateUpsert() error = %v", err)
	}
	if req.Code != "pro" || req.Name != "Pro" || req.Currency != "USD" {
		t.Fatalf("validateUpsert() normalization = %#v", req)
	}
}

func TestValidateUpsertRejectsInvalidAmount(t *testing.T) {
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
