package config

import (
	"fmt"
	"net/netip"
	"strings"

	"github.com/google/uuid"
)

func (c Config) Validate() error {
	return c.Polar.Validate()
}

// ValidateForServer requires only the secrets and connections used by the API.
func (c Config) ValidateForServer() error {
	return requireSettings(map[string]string{
		"DATABASE_URL":            c.DatabaseURL,
		"REDIS_URL":               c.RedisURL,
		"FREESWITCH_ESL_PASSWORD": c.FreeSWITCHESLPassword,
		"ENCRYPTION_KEY":          c.EncryptionKey,
		"TURN_AUTH_SECRET":        c.TURNAuthSecret,
		"KEYCLOAK_CLIENT_SECRET":  c.Keycloak.ClientSecret,
	})
}

// ValidateForWorker excludes Keycloak and TURN secrets not used by background jobs.
func (c Config) ValidateForWorker() error {
	return requireSettings(map[string]string{
		"DATABASE_URL":            c.DatabaseURL,
		"REDIS_URL":               c.RedisURL,
		"NATS_URL":                c.NATSURL,
		"FREESWITCH_ESL_PASSWORD": c.FreeSWITCHESLPassword,
		"MEDIA_CONTROL_TOKEN":     c.MediaControlToken,
		"ENCRYPTION_KEY":          c.EncryptionKey,
	})
}

// ValidateForBackoffice keeps the admin service independent of telephony and AI secrets.
func (c Config) ValidateForBackoffice() error {
	return requireSettings(map[string]string{
		"DATABASE_URL":           c.DatabaseURL,
		"ENCRYPTION_KEY":         c.EncryptionKey,
		"KEYCLOAK_CLIENT_SECRET": c.Keycloak.ClientSecret,
	})
}

func requireSettings(values map[string]string) error {
	for _, key := range []string{
		"DATABASE_URL",
		"REDIS_URL",
		"NATS_URL",
		"FREESWITCH_ESL_PASSWORD",
		"MEDIA_CONTROL_TOKEN",
		"ENCRYPTION_KEY",
		"TURN_AUTH_SECRET",
		"KEYCLOAK_CLIENT_SECRET",
	} {
		if value, ok := values[key]; ok && strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s is required", key)
		}
	}
	return nil
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
