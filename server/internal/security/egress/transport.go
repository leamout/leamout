package egress

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

// NewTransport enforces public destination addresses at connect time.
//
// Proxies are deliberately disabled. A proxy resolves the target independently,
// so validating only the address dialed by this process would validate the proxy
// rather than the actual webhook destination. Explicit proxy support requires a
// trusted proxy that independently enforces the same destination restrictions.
func NewTransport() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil

	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport.DialContext = func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("split webhook destination address: %w", err)
		}
		addresses, err := ResolveHost(ctx, host)
		if err != nil {
			return nil, err
		}
		var lastErr error
		for _, ip := range addresses {
			conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if dialErr == nil {
				return conn, nil
			}
			lastErr = dialErr
		}
		if lastErr == nil {
			lastErr = fmt.Errorf("webhook destination resolved to no dialable addresses")
		}
		return nil, fmt.Errorf("dial webhook destination: %w", lastErr)
	}
	transport.MaxIdleConns = 20
	transport.MaxIdleConnsPerHost = 2
	transport.IdleConnTimeout = 30 * time.Second
	transport.ResponseHeaderTimeout = 10 * time.Second
	return transport
}

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

// ResolveHost refuses mixed public/private DNS results and binds the validated
// addresses to the actual dial, preventing DNS rebinding between check and use.
func ResolveHost(ctx context.Context, host string) ([]netip.Addr, error) {
	if parsed, err := netip.ParseAddr(host); err == nil {
		if err := ValidateAddress(parsed); err != nil {
			return nil, err
		}
		return []netip.Addr{parsed}, nil
	}

	addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("resolve webhook destination: %w", err)
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("webhook destination resolved to no addresses")
	}
	for _, address := range addresses {
		if err := ValidateAddress(address); err != nil {
			return nil, err
		}
	}
	return addresses, nil
}

// ValidateAddress excludes non-public ranges, including IPv4-mapped IPv6 and
// carrier-grade NAT space.
func ValidateAddress(address netip.Addr) error {
	address = address.Unmap()
	if !address.IsValid() ||
		!address.IsGlobalUnicast() ||
		address.IsPrivate() ||
		address.IsLoopback() ||
		address.IsLinkLocalUnicast() ||
		address.IsLinkLocalMulticast() ||
		address.IsMulticast() ||
		address.IsUnspecified() ||
		isCarrierGradeNAT(address) {
		return fmt.Errorf("webhook destination resolves to a non-public address")
	}
	return nil
}

func isCarrierGradeNAT(address netip.Addr) bool {
	return address.Is4() && netip.MustParsePrefix("100.64.0.0/10").Contains(address)
}
