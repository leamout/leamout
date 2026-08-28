package sip_domains

import "testing"

func TestNormalizeDomain(t *testing.T) {
	got, err := normalizeDomain("  SIP.Example.COM. ")
	if err != nil || got != "sip.example.com" {
		t.Fatalf("normalizeDomain() = %q, %v", got, err)
	}
	for _, value := range []string{"localhost", "127.0.0.1", "-bad.example", "bad_.example"} {
		if _, err := normalizeDomain(value); err == nil {
			t.Errorf("normalizeDomain(%q) accepted invalid domain", value)
		}
	}
}
