package config

import (
	"fmt"
	"net/netip"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/platform/deployment"
)

func (c Config) Validate() error {
	if c.Commercial.Mode != deployment.SelfHosted {
		if err := c.Polar.Validate(); err != nil {
			return err
		}
	}
	return c.Commercial.Validate()
}

func (c PolarConfig) Validate() error {
	if c.Environment != "production" && c.Environment != "sandbox" {
		return fmt.Errorf("POLAR_ENVIRONMENT must be production or sandbox")
	}
	seen := make(map[uuid.UUID]bool)
	for _, value := range []string{
		c.DeveloperProductID,
		c.ProProductID,
		c.ScaleProductID,
	} {
		id, err := uuid.Parse(value)
		if err != nil || id == uuid.Nil || seen[id] {
			return fmt.Errorf("polar product IDs must be distinct nonempty UUIDs")
		}
		seen[id] = true
	}
	return nil
}

func (c Config) TrustedProxyPrefixes() ([]netip.Prefix, error) {
	result := make([]netip.Prefix, 0, len(c.TrustedProxyCIDRs))
	for _, value := range c.TrustedProxyCIDRs {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return nil, fmt.Errorf("parse trusted proxy CIDR %q: %w", value, err)
		}
		result = append(result, prefix.Masked())
	}
	return result, nil
}
