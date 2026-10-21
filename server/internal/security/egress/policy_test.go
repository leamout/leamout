package egress

import (
 "context"
 "net/netip"
 "testing"
)

func TestDefaultPolicyBlocksPrivateSIPDestination(t *testing.T) {
 private := netip.MustParseAddr("172.30.0.50")
 if err := ValidateAddress(private); err == nil { t.Fatal("unexpected private egress") }
 if _, err := ResolveHost(context.Background(), "127.0.0.1"); err == nil { t.Fatal("unexpected loopback egress") }
 fixture := Policy{AdditionalAllowedIPs: []netip.Addr{private}}
 if err := fixture.ValidateAddress(private); err != nil { t.Fatal(err) }
 if err := ValidateAddress(private); err == nil { t.Fatal("fixture policy modified global restrictions") }
 if err := fixture.ValidateAddress(netip.MustParseAddr("169.254.169.254")); err == nil { t.Fatal("unexpected metadata access") }
}
