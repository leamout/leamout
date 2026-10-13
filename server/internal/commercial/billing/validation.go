package billing

import (
	"net/url"
	"strings"

	"github.com/leamout/leamout/server/pkg/apperror"
)

// validateRedirects prevents authenticated clients from using Polar checkout
// and portal links as open redirects to attacker-controlled websites.
func validateRedirects(frontendURL string, values ...string) error {
	trusted, err := url.Parse(strings.TrimSpace(frontendURL))
	if err != nil ||
		trusted.Host == "" ||
		(trusted.Scheme != "https" && trusted.Scheme != "http") ||
		trusted.User != nil {
		return apperror.NewServiceUnavailable("billing frontend origin is not configured", nil)
	}
	for _, raw := range values {
		target, parseErr := url.Parse(strings.TrimSpace(raw))
		if parseErr != nil ||
			!target.IsAbs() ||
			target.User != nil ||
			target.Scheme != trusted.Scheme ||
			!strings.EqualFold(target.Host, trusted.Host) ||
			strings.HasPrefix(target.Path, "//") {
			return apperror.NewBadRequest("billing redirect URL must use the configured frontend origin")
		}
	}
	return nil
}
