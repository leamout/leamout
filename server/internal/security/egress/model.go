package egress

import "net/netip"

// Policy is scoped to an outbound connection path. Zero value enforces
// public-only destinations; additional explicit IPs are only for callers
// which independently authorize them. Never apply a relaxed policy to
// webhook destinations or arbitrary user-supplied URLs.
type Policy struct {
	AdditionalAllowedIPs []netip.Addr
}
