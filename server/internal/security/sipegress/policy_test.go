package sipegress

import (
	"context"
	"strings"
	"testing"
)

func TestAcceptanceCarrierPolicy(t *testing.T) {
	t.Setenv("APP_ENV", "acceptance")
	for _, host := range []string{"byoc-v1-carrier", "voice-v1-carrier", "voice-agent-v1-carrier", "graceful-drain-carrier"} {
		if !IsAcceptanceFixture(host) {
			t.Fatalf("%q must be isolated acceptance fixture", host)
		}
		addresses, err := ResolveHost(context.Background(), host)
		if err != nil || len(addresses) != 1 || addresses[0].String() != "172.30.0.50" {
			t.Fatalf("%q = %v, %v; want only acceptance carrier IP", host, addresses, err)
		}
	}
	for _, host := range []string{"127.0.0.1", "10.0.0.1", "169.254.169.254"} {
		if _, err := ResolveHost(context.Background(), host); err == nil {
			t.Fatalf("private SIP endpoint %q was accepted", host)
		}
	}
}

func TestProductionRejectsAcceptanceCarrier(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	if IsAcceptanceFixture("byoc-v1-carrier") {
		t.Fatal("acceptance fixture enabled in production")
	}
	if err := ValidateHost("127.0.0.1"); err == nil {
		t.Fatal("private SIP endpoint was accepted")
	}
	if _, err := ResolveHost(context.Background(), "127.0.0.1"); err == nil || !strings.Contains(err.Error(), "non-public") {
		t.Fatalf("private SIP egress error = %v", err)
	}
}
