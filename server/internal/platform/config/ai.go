package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

func (c Config) PlatformAICredentials() (map[string]string, error) {
	raw := strings.TrimSpace(c.AIPlatformCredentials)
	if raw == "" {
		return map[string]string{}, nil
	}

	var values map[string]string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, fmt.Errorf("parse AI_PLATFORM_CREDENTIALS: %w", err)
	}

	result := make(map[string]string, len(values))
	for provider, secret := range values {
		provider = strings.TrimSpace(provider)
		secret = strings.TrimSpace(secret)
		if provider == "" {
			return nil, fmt.Errorf("parse AI_PLATFORM_CREDENTIALS: provider name is required")
		}
		if secret == "" {
			return nil, fmt.Errorf("parse AI_PLATFORM_CREDENTIALS: credential for %q is empty", provider)
		}
		result[provider] = secret
	}
	return result, nil
}
