package config

import "testing"

func TestValidateDeploymentModes(t *testing.T) {
	for _, mode := range []DeploymentMode{DeploymentModeCloud, DeploymentModeSelfHosted} {
		cfg := Config{DeploymentMode: mode}
		if err := cfg.ValidateDeployment(); err != nil {
			t.Fatalf("ValidateDeployment(%q) error = %v", mode, err)
		}
	}
	for _, mode := range []DeploymentMode{"", "hosted", "SELF-HOSTED"} {
		cfg := Config{DeploymentMode: mode}
		if err := cfg.ValidateDeployment(); err == nil {
			t.Fatalf("ValidateDeployment(%q) unexpectedly succeeded", mode)
		}
	}
}

func TestSelfHostedRejectsCloudOnlyConfiguration(t *testing.T) {
	tests := []struct {
		name string
		set  func(*Config)
	}{
		{"DIDWW", func(c *Config) { c.DIDWW.APIKey = "secret" }},
		{"CommPeak", func(c *Config) { c.CommPeak.Authorization = "secret" }},
		{"Stripe", func(c *Config) { c.Stripe.SecretKey = "secret" }},
		{"Paystack", func(c *Config) { c.Paystack.SecretKey = "secret" }},
		{"managed SIP", func(c *Config) { c.ManagedSIP.AdmissionSecret = "secret" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{DeploymentMode: DeploymentModeSelfHosted}
			tt.set(&cfg)
			if err := cfg.ValidateDeployment(); err == nil {
				t.Fatal("ValidateDeployment() accepted Cloud-only configuration")
			}
		})
	}
}

func TestNormalizeDefaultsDevelopmentToSelfHosted(t *testing.T) {
	cfg := Config{AppEnv: " development "}
	cfg.normalize()
	if cfg.DeploymentMode != DeploymentModeSelfHosted {
		t.Fatalf("DeploymentMode = %q, want %q", cfg.DeploymentMode, DeploymentModeSelfHosted)
	}
}

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
