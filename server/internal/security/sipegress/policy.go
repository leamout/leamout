package sipegress

import (
	"context"
	"fmt"
	"net/netip"
	"os"

	"github.com/leamout/leamout/server/internal/security/egress"
)

const acceptanceCarrierIP = "172.30.0.50"

var acceptanceCarriers = map[string]struct{}{
	"byoc-v1-carrier": {},
	"voice-v1-carrier": {},
	"voice-agent-v1-carrier": {},
	"graceful-drain-carrier": {},
}

// ResolveHost is the single public BYOC endpoint policy for health probes and
// call origination. The only exception is a fixed fixture address in disposable
// acceptance stacks; no arbitrary private IPs or DNS overrides are supported.
func ResolveHost(ctx context.Context, host string) ([]netip.Addr, error) {
	if _, ok := acceptanceCarriers[host]; ok && os.Getenv("APP_ENV") == "acceptance" {
		return []netip.Addr{netip.MustParseAddr(acceptanceCarrierIP)}, nil
	}
	if os.Getenv("APP_ENV") == "acceptance" {
		// Refuse accidental private routing even in acceptance mode.
		if ip, err := netip.ParseAddr(host); err == nil {
			if err := egress.ValidateAddress(ip); err != nil {
				return nil, err
			}
		}
	}
	addresses, err := egress.ResolveHost(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("SIP trunk egress: %w", err)
	}
	return addresses, nil
}

// ValidateHost rejects private literals and local-only names at API write time.
// The acceptance fixture is restricted to its four names and exact test address.
func ValidateHost(host string) error {
	if os.Getenv("APP_ENV") == "acceptance" {
		if _, ok := acceptanceCarriers[host]; ok {
			return nil
		}
	}
	if address, err := netip.ParseAddr(host); err == nil {
		return egress.ValidateAddress(address)
	}
	if host == "localhost" {
		return fmt.Errorf("SIP endpoint must be publicly routable")
	}
	return nil
}

// IsAcceptanceFixture is intended only for test-policy regression assertions.
func IsAcceptanceFixture(host string) bool {
	_, ok := acceptanceCarriers[host]
	return ok && os.Getenv("APP_ENV") == "acceptance"
}

