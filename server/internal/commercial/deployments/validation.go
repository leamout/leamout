package deployments

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrOrganizationRequired   = errors.New("organization_id is required")
	ErrLicenseRequired        = errors.New("license_id is required")
	ErrDeploymentIDRequired   = errors.New("deployment_id is required")
	ErrInvalidDeploymentName  = errors.New("deployment name must not be blank")
	ErrActivationTimeRequired = errors.New("activation time is required")
)

func ValidateActivate(request ActivateRequest) error {
	if request.OrganizationID == uuid.Nil {
		return ErrOrganizationRequired
	}
	if request.LicenseID == uuid.Nil {
		return ErrLicenseRequired
	}
	if strings.TrimSpace(request.DeploymentID) == "" {
		return ErrDeploymentIDRequired
	}
	if request.Name != nil && strings.TrimSpace(*request.Name) == "" {
		return ErrInvalidDeploymentName
	}
	if request.At.IsZero() {
		return ErrActivationTimeRequired
	}
	return nil
}
