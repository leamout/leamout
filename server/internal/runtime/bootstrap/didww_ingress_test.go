package bootstrap

import (
	"net/netip"
	"testing"
)

func TestParseSourceCIDRsNormalizesAndDeduplicates(t *testing.T) {
	items, err := parseSourceCIDRs([]string{"203.0.113.9/24", "203.0.113.0/24", "2001:db8::1/64"})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("count = %d; want 2", len(items))
	}
	if items[0] != netip.MustParsePrefix("203.0.113.0/24") {
		t.Fatalf("first CIDR = %s", items[0])
	}
	if items[1] != netip.MustParsePrefix("2001:db8::/64") {
		t.Fatalf("second CIDR = %s", items[1])
	}
}

func TestParseSIPEndpoint(t *testing.T) {
	endpoint, err := parseSIPEndpoint("sip.didww.example:5061/tls")
	if err != nil {
		t.Fatal(err)
	}
	if endpoint.Host != "sip.didww.example" || endpoint.Port != 5061 || endpoint.Transport != "tls" {
		t.Fatalf("endpoint = %+v", endpoint)
	}
}

func TestParseSIPEndpointDefaults(t *testing.T) {
	endpoint, err := parseSIPEndpoint("sip.didww.example")
	if err != nil {
		t.Fatal(err)
	}
	if endpoint.Host != "sip.didww.example" || endpoint.Port != 5060 || endpoint.Transport != "udp" {
		t.Fatalf("endpoint = %+v", endpoint)
	}
}

func TestDIDWWExternalReferenceIsDeploymentScoped(t *testing.T) {
	got := didwwExternalReference("deployment-123")
	if got != "leamout:deployment-123:managed-ingress" {
		t.Fatalf("external reference = %q", got)
	}
}
