package email

import (
	"fmt"
	"net/url"
	"regexp"
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
		if !codePattern.MatchString(data.Code) {
			return fmt.Errorf("invalid verification code")
		}
	case "organization-invitation":
		if !validURL(data.AcceptURL) || data.Organization == "" {
			return fmt.Errorf("invalid invitation data")
		}
	case "password-changed":
		if !validURL(data.RecoveryURL) || data.ChangedAt == "" {
			return fmt.Errorf("invalid password notification data")
		}
	}
	return nil
}

func validURL(value string) bool {
	u, err := url.Parse(value)
	return err == nil && u.Scheme == "https" && u.Host != "" && u.User == nil
}
