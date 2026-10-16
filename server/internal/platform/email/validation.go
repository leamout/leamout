package email

import (
	"fmt"
	"net/netip"
	"net/url"
	"regexp"
	"strings"
)

var templateNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func validateData(name string, data Data) error {
	if !templateNamePattern.MatchString(name) {
		return fmt.Errorf("invalid template name")
	}
	if data.ExpiresAt.IsZero() {
		return fmt.Errorf("email expiry is required")
	}
	switch name {
	case "organization-invitation":
		if !ValidURL(data.AcceptURL) || data.Organization == "" {
			return fmt.Errorf("invalid invitation data")
		}
	default:
		return fmt.Errorf("unsupported email template")
	}
	return nil
}

// ValidURL accepts HTTPS links and HTTP links to localhost or loopback addresses.
func ValidURL(value string) bool {
	u, err := url.Parse(value)
	if err != nil || u.Hostname() == "" || u.User != nil {
		return false
	}

	if u.Scheme == "https" {
		return true
	}

	if u.Scheme != "http" {
		return false
	}

	if strings.EqualFold(u.Hostname(), "localhost") {
		return true
	}

	address, err := netip.ParseAddr(u.Hostname())
	return err == nil && address.IsLoopback()
}
