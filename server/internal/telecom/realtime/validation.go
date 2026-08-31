package realtime

import (
	"fmt"
	"net/url"
	"strings"
	"time"
)

const maximumCredentialTTL = 24 * time.Hour

func validateConfig(config Config) error {
	if len(config.AuthSecret) < 32 {
		return fmt.Errorf("TURN authentication secret must be at least 32 bytes")
	}
	if len(config.URLs) == 0 {
		return fmt.Errorf("at least one TURN or STUN URL is required")
	}
	for _, rawURL := range config.URLs {
		if !validICEURL(strings.TrimSpace(rawURL)) {
			return fmt.Errorf("invalid TURN or STUN URL %q", rawURL)
		}
	}
	if config.TTL <= 0 || config.TTL > maximumCredentialTTL {
		return fmt.Errorf("TURN credential TTL must be between 1ns and %s", maximumCredentialTTL)
	}
	if config.IssueLimit <= 0 {
		return fmt.Errorf("TURN credential issue limit must be positive")
	}
	if config.IssueWindow <= 0 {
		return fmt.Errorf("TURN credential issue window must be positive")
	}
	return nil
}

func validICEURL(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Fragment != "" {
		return false
	}
	switch parsed.Scheme {
	case "stun", "stuns", "turn", "turns":
	default:
		return false
	}

	// RFC 7064/7065 URLs normally parse the host and port into Opaque because
	// they do not use // after the scheme. Also accept a conventional Host form
	// while requiring an actual relay target in either representation.
	target := parsed.Opaque
	if target == "" {
		target = parsed.Host
	}
	target, _, _ = strings.Cut(target, "?")
	return strings.TrimSpace(target) != ""
}
