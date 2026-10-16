package entitlements

import "testing"

func TestCapabilityValidation(t *testing.T) {
	tests := []struct {
		name       string
		capability Capability
		valid      bool
	}{
		{name: "advanced RBAC", capability: CapabilityAdvancedRBAC, valid: true},
		{name: "retention policies", capability: CapabilityRetentionPolicies, valid: true},
		{name: "private networking", capability: CapabilityPrivateNetworking, valid: true},
		{name: "unknown", capability: "enterprise_runtime", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.capability.IsValid(); got != test.valid {
				t.Fatalf("IsValid() = %t, want %t", got, test.valid)
			}
		})
	}
}

func TestNormalizeCapabilityTrimsWhitespace(t *testing.T) {
	got, err := normalizeCapability("  advanced_rbac  ")
	if err != nil {
		t.Fatalf("normalizeCapability() error = %v", err)
	}
	if got != CapabilityAdvancedRBAC {
		t.Fatalf("normalizeCapability() = %q, want %q", got, CapabilityAdvancedRBAC)
	}
}
