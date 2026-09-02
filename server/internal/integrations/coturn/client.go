package coturn

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

const defaultHealthTimeout = 2 * time.Second

type Config struct {
	Address string
	Timeout time.Duration
}

type Client struct {
	address string
	dialer  *net.Dialer
}

func DefaultConfig(address string) Config {
	return Config{
		Address: strings.TrimSpace(address),
		Timeout: defaultHealthTimeout,
	}
}

func New(config Config) (*Client, error) {
	config.Address = strings.TrimSpace(config.Address)
	if config.Address == "" {
		return nil, fmt.Errorf("Coturn health address is required")
	}

	host, port, err := net.SplitHostPort(config.Address)
	if err != nil || host == "" || port == "" {
		return nil, fmt.Errorf("invalid Coturn health address %q", config.Address)
	}
	if config.Timeout <= 0 {
		return nil, fmt.Errorf("Coturn health timeout must be positive")
	}

	return &Client{
		address: config.Address,
		dialer:  &net.Dialer{Timeout: config.Timeout},
	}, nil
}

// HealthCheck verifies that the Coturn listener is reachable over the internal
// control network. Credential correctness is exercised separately through the
// real TURN acceptance flow.
func (c *Client) HealthCheck(ctx context.Context) error {
	if c == nil || c.dialer == nil || c.address == "" {
		return fmt.Errorf("Coturn client is not initialized")
	}
	if ctx == nil {
		return fmt.Errorf("context is required")
	}

	conn, err := c.dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return fmt.Errorf("connect Coturn %s: %w", c.address, err)
	}
	if err := conn.Close(); err != nil {
		return fmt.Errorf("close Coturn health connection: %w", err)
	}
	return nil
}
