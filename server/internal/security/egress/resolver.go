package egress

import (
	"context"
	"fmt"
	"net"
	"net/netip"
)

// ResolveHost refuses mixed public/private DNS results and binds the validated
// addresses to the actual dial, preventing DNS rebinding between check and use.
func ResolveHost(ctx context.Context, host string) ([]netip.Addr, error) {
	return (Policy{}).ResolveHost(ctx, host)
}

// ResolveHost verifies all answers before returning any IP. Callers must dial
// the returned IP itself, not the original hostname, to prevent DNS rebinding.
func (p Policy) ResolveHost(ctx context.Context, host string) ([]netip.Addr, error) {
	if parsed, err := netip.ParseAddr(host); err == nil {
		if err := p.ValidateAddress(parsed); err != nil {
			return nil, err
		}
		return []netip.Addr{parsed}, nil
	}

	addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("resolve destination: %w", err)
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("destination resolved to no addresses")
	}
	for _, address := range addresses {
		if err := p.ValidateAddress(address); err != nil {
			return nil, err
		}
	}
	return addresses, nil
}

