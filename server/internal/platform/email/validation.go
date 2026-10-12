package email

import (
	"fmt"
	"net/netip"
	"net/url"
	"regexp"
	"strings"
)

var codePattern = regexp.MustCompile(`^[0-9]{6}$`)
var templateNamePattern = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func validateData(name string, data Data) error {
	if !templateNamePattern.MatchString(name) {
		return fmt.Errorf("invalid template name")
	}
	if data.ExpiresAt.IsZero() {
		return fmt.Errorf("email expiry is required")
	}
	switch name {
	case "verification-code":
		if data.Purpose != "" && data.Purpose != "verification" && data.Purpose != "password-recovery" {
			return fmt.Errorf("invalid verification purpose")
		}

		if !codePattern.MatchString(data.Code) {
			return fmt.Errorf("invalid verification code")
		}
	case "organization-invitation":
		if !ValidURL(data.AcceptURL) || data.Organization == "" {
			return fmt.Errorf("invalid invitation data")
		}
	case "password-changed":
		if !ValidURL(data.RecoveryURL) || data.ChangedAt == "" {
			return fmt.Errorf("invalid password notification data")
		}
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
