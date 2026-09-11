package payments

import "testing"

func TestValidateCheckoutNormalizesCurrency(t *testing.T) {
	request, err := ValidateCheckout(CheckoutRequest{
		Reference: " invoice-1 ", AmountMinor: 2500, Currency: " ghs ", Email: " buyer@example.com ",
		CallbackURL: "https://console.leamout.com/billing/return",
	})
	if err != nil {
		t.Fatal(err)
	}
	if request.Reference != "invoice-1" || request.Currency != "GHS" || request.Email != "buyer@example.com" {
		t.Fatalf("normalized request = %+v", request)
	}
}

func TestValidateCheckoutRejectsUntrustedCommercialInputs(t *testing.T) {
	tests := []CheckoutRequest{
		{AmountMinor: 1, Currency: "GHS", Email: "buyer@example.com"},
		{Reference: "invoice-1", AmountMinor: 0, Currency: "GHS", Email: "buyer@example.com"},
		{Reference: "invoice-1", AmountMinor: 1, Currency: "cedi", Email: "buyer@example.com"},
		{Reference: "invoice-1", AmountMinor: 1, Currency: "GHS"},
		{Reference: "invoice-1", AmountMinor: 1, Currency: "GHS", Email: "buyer@example.com", CallbackURL: "http://example.com/return"},
	}
	for _, request := range tests {
		if _, err := ValidateCheckout(request); err == nil {
			t.Fatalf("expected validation error for %+v", request)
		}
	}
}
