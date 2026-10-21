package egress

import (
	"context"
	"fmt"
	"net"
	"net/http"
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
