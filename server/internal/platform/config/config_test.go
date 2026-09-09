package config

import "testing"

func TestNormalizeCommPeakConfig(t *testing.T) {
	cfg := Config{
		CommPeak: CommPeakConfig{
			Authorization: "  Bearer test-token  ",
			APIBaseURL:    "  https://api.commpeak.com/  ",
		},
	}

	cfg.normalize()

	if got, want := cfg.CommPeak.Authorization, "Bearer test-token"; got != want {
		t.Fatalf("CommPeak API authorization = %q, want %q", got, want)
	}
	if got, want := cfg.CommPeak.APIBaseURL, "https://api.commpeak.com"; got != want {
		t.Fatalf("CommPeak API base URL = %q, want %q", got, want)
	}
}

func TestNormalizePaymentProviderConfig(t *testing.T) {
	cfg := Config{
		Stripe:   StripeConfig{SecretKey: "  sk_test  ", WebhookSecret: "  whsec_test  ", APIBaseURL: " https://api.stripe.test/v1/ "},
		Paystack: PaystackConfig{SecretKey: "  sk_test_paystack  ", APIBaseURL: " https://api.paystack.test/ "},
	}

	cfg.normalize()

	if cfg.Stripe.SecretKey != "sk_test" || cfg.Stripe.WebhookSecret != "whsec_test" || cfg.Stripe.APIBaseURL != "https://api.stripe.test/v1" {
		t.Fatalf("Stripe config was not normalized: %+v", cfg.Stripe)
	}
	if cfg.Paystack.SecretKey != "sk_test_paystack" || cfg.Paystack.APIBaseURL != "https://api.paystack.test" {
		t.Fatalf("Paystack config was not normalized: %+v", cfg.Paystack)
	}
}
