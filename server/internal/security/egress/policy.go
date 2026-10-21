package egress

import (
	"fmt"
	"net/netip"
	"net/url"
	"strings"
)

// ParseHTTPS rejects non-HTTPS, ambiguous, or credential-bearing webhook URLs.
// Literal IP addresses are checked here; DNS results are rechecked on each dial.
func ParseHTTPS(raw string) (*url.URL, error) {
	endpoint, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || endpoint.Scheme != "https" || endpoint.Hostname() == "" || endpoint.Opaque != "" {
		return nil, fmt.Errorf("webhook destination must be an absolute HTTPS URL")
	}
	if endpoint.User != nil {
		return nil, fmt.Errorf("webhook destination cannot contain userinfo")
	}
	if endpoint.Fragment != "" {
		return nil, fmt.Errorf("webhook destination cannot contain a fragment")
	}
	if address, err := netip.ParseAddr(endpoint.Hostname()); err == nil {
		if err := ValidateAddress(address); err != nil {
			return nil, err
		}
	}
	return endpoint, nil
}

// ValidateAddress excludes non-public ranges, including IPv4-mapped IPv6 and
// carrier-grade NAT space.
func ValidateAddress(address netip.Addr) error {
	return (Policy{}).ValidateAddress(address)
}

func (p Policy) ValidateAddress(address netip.Addr) error {
	address = address.Unmap()
	for _, allowed := range p.AdditionalAllowedIPs {
		if allowed.IsValid() && address == allowed.Unmap() {
			return nil
		}
	}
	if !address.IsValid() ||
		!address.IsGlobalUnicast() ||
		address.IsPrivate() ||
		address.IsLoopback() ||
		address.IsLinkLocalUnicast() ||
		address.IsLinkLocalMulticast() ||
		address.IsMulticast() ||
		address.IsUnspecified() ||
		isCarrierGradeNAT(address) {
		return fmt.Errorf("destination resolves to a non-public address")
	}
	return nil
}

func isCarrierGradeNAT(address netip.Addr) bool {
	return address.Is4() && netip.MustParsePrefix("100.64.0.0/10").Contains(address)
}
