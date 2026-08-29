package licensing

import (
	"errors"
	"strings"
)

var (
	ErrCustomerRequired       = errors.New("customer_id is required")
	ErrInvalidDeploymentLimit = errors.New("max_deployments must be greater than zero")
	ErrInvalidExpiration      = errors.New("expires_at must be after issued_at")
)

// Validate checks invariants required before a license is persisted or issued.
func Validate(license License) error {
	if strings.TrimSpace(license.CustomerID) == "" {
		return ErrCustomerRequired
	}
	if license.MaxDeployments <= 0 {
		return ErrInvalidDeploymentLimit
	}
	if license.ExpiresAt != nil && !license.ExpiresAt.After(license.IssuedAt) {
		return ErrInvalidExpiration
	}
	return nil
}
