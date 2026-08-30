package licensing

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

var (
	ErrLicenseNotActive = errors.New("license is not active")
	ErrLicenseExpired   = errors.New("license is expired")
)

func validateIssueRequest(req IssueRequest) error {
	if req.OrganizationID == uuid.Nil {
		return fmt.Errorf("organization_id is required")
	}
	if req.LicenseID == uuid.Nil {
		return fmt.Errorf("license_id is required")
	}
	if req.At.IsZero() {
		return fmt.Errorf("evaluation time is required")
	}
	return nil
}

func validateActiveLicense(license sqlc.License, at time.Time) error {
	if license.Status != string(StatusActive) || !license.IssuedAt.Valid || at.Before(license.IssuedAt.Time) {
		return ErrLicenseNotActive
	}
	if license.ExpiresAt.Valid && !at.Before(license.ExpiresAt.Time) {
		return ErrLicenseExpired
	}
	if license.MaxDeployments <= 0 {
		return fmt.Errorf("license has invalid max_deployments")
	}
	if license.SigningKeyID == nil || *license.SigningKeyID == "" {
		return fmt.Errorf("license has no signing key")
	}
	return nil
}
