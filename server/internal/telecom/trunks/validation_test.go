package trunks

import "testing"

func TestNormalizeHost(t *testing.T) {
	got, err := normalizeHost(" SIP.Example.COM. ")
	if err != nil || got != "sip.example.com" {
		t.Fatalf("normalizeHost() = %q, %v", got, err)
	}
	for _, value := range []string{"", "bad host", "-example.com", "example..com"} {
		if _, err := normalizeHost(value); err == nil {
			t.Errorf("normalizeHost(%q) accepted invalid host", value)
		}
	}
}

func TestEndpointRanges(t *testing.T) {
	if validatePort(1) != nil || validatePort(65535) != nil {
		t.Fatal("valid port rejected")
	}
	if validatePort(0) == nil || validatePort(65536) == nil {
		t.Fatal("invalid port accepted")
	}
	if validatePriority(-1) == nil {
		t.Fatal("negative priority accepted")
	}
	if validateWeight(0) == nil {
		t.Fatal("zero weight accepted")
	}
}

func TestNormalizeChoices(t *testing.T) {
	got, err := normalizeChoice(" OUTBOUND ", directions, "direction")
	if err != nil || got != "outbound" {
		t.Fatalf("normalizeChoice() = %q, %v", got, err)
	}
	if _, err := normalizeChoice("sideways", directions, "direction"); err == nil {
		t.Fatal("invalid direction accepted")
	}
}

func TestNormalizeProvisioningMode(t *testing.T) {
	for _, value := range []ProvisioningMode{"", ProvisioningModeBYOC, " BYOC "} {
		got, err := normalizeProvisioningMode(value)
		if err != nil || got != ProvisioningModeBYOC {
			t.Fatalf("normalizeProvisioningMode(%q) = %q, %v; want byoc", value, got, err)
		}
	}

	for _, value := range []ProvisioningMode{ProvisioningModeManaged, " MANAGED ", "other"} {
		if _, err := normalizeProvisioningMode(value); err == nil {
			t.Fatalf("normalizeProvisioningMode(%q) accepted non-BYOC customer trunk mode", value)
		}
	}
}
