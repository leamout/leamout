package config

import (
	"os/exec"
	"reflect"
	"strings"
	"testing"
)

func TestSharedConfigExcludesCloudCapabilities(t *testing.T) {
	cfg := Config{AppEnv: "test"}
	cloud := CloudConfig{}

	// These assignments are the compile-time contract: Cloud embeds the shared
	// runtime configuration, while provider settings are only available through
	// the Cloud-specific type.
	cloud.Config = cfg
	cloud.DIDWW.APIKey = "cloud-only"
	if cloud.AppEnv != "test" || cloud.DIDWW.APIKey != "cloud-only" {
		t.Fatal("Cloud configuration does not preserve the shared/configured capability boundary")
	}
	sharedType := reflect.TypeOf(cfg)
	for _, field := range []string{"DIDWW", "CommPeak", "ManagedSIP", "Stripe", "Paystack", "OperatorAPISecret"} {
		if _, exists := sharedType.FieldByName(field); exists {
			t.Errorf("shared configuration exposes Cloud-only field %s", field)
		}
	}
}

func TestDeploymentImportBoundaries(t *testing.T) {
	tests := []struct {
		name      string
		packages  []string
		forbidden []string
	}{
		{
			name:     "self-hosted",
			packages: []string{"../../../cmd/selfhosted", "../../../cmd/selfhosted-worker"},
			forbidden: []string{
				"/internal/runtime/cloud",
				"/internal/runtime/cloudworker",
				"/internal/commercial",
				"/internal/integrations/payments/",
				"/internal/integrations/carriers/",
				"/internal/platform/provider_diagnostics",
				"/internal/telecom/edge",
				"/internal/telecom/wholesale",
			},
		},
		{
			name:     "cloud",
			packages: []string{"../../../cmd/cloud", "../../../cmd/cloud-worker"},
			forbidden: []string{
				"/internal/runtime/selfhosted",
				"/internal/runtime/selfhostedworker",
				"/internal/runtime/leamout",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{"list", "-deps"}, tt.packages...)
			output, err := exec.Command("go", args...).CombinedOutput()
			if err != nil {
				t.Fatalf("inspect deployment dependency graph: %v\n%s", err, output)
			}
			dependencies := string(output)
			for _, forbidden := range tt.forbidden {
				if strings.Contains(dependencies, forbidden) {
					t.Errorf("dependency graph contains forbidden package boundary %q", forbidden)
				}
			}
		})
	}
}

func TestNormalizeCommPeakConfig(t *testing.T) {
	cfg := CloudConfig{
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
	cfg := CloudConfig{
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
