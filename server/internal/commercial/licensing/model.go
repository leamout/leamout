package licensing

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Status describes whether a commercial license may authorize self-hosted deployments.
type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusExpired   Status = "expired"
	StatusRevoked   Status = "revoked"
)

// DeploymentStatus describes whether an installation currently consumes a license slot.
type DeploymentStatus string

const (
	DeploymentStatusActive      DeploymentStatus = "active"
	DeploymentStatusDeactivated DeploymentStatus = "deactivated"
)

const MaxDeploymentsEntitlement = "max.deployments"

var (
	ErrLicenseNotFound            = errors.New("license not found")
	ErrLicenseUnavailable         = errors.New("license is unavailable")
	ErrCommercialStateUnavailable = errors.New("commercial state does not allow license creation")
	ErrSubscriptionUnavailable    = errors.New("current subscription is unavailable for licensing")
	ErrOrganizationIDRequired     = errors.New("organization id is required")
	ErrLicenseIDRequired          = errors.New("license id is required")
	ErrSigningKeyRequired         = errors.New("signing key id is required before license activation")
	ErrInvalidSigningKey          = errors.New("signing key id must not contain whitespace")
	ErrSigningKeyUnavailable      = errors.New("license signing key is unavailable")
	ErrUnsupportedLicenseVersion  = errors.New("unsupported signed license version")
	ErrUnsupportedAlgorithm       = errors.New("unsupported license signature algorithm")
	ErrMalformedArtifact          = errors.New("malformed signed license artifact")
	ErrInvalidSignature           = errors.New("invalid signed license signature")
	ErrArtifactExpired            = errors.New("signed license artifact has expired")
	ErrArtifactNotYetValid        = errors.New("signed license artifact is not yet valid")
	ErrDeploymentMismatch         = errors.New("signed license is bound to another deployment")
	ErrInvalidStatus              = errors.New("invalid license status")
	ErrInvalidTransition          = errors.New("invalid license status transition")
	ErrInvalidDeploymentLimit     = errors.New("max deployments must be greater than zero")
	ErrInvalidExpiration          = errors.New("license expiration must be after issuance")
	ErrDeploymentIDRequired       = errors.New("deployment id is required")
	ErrInvalidDeploymentID        = errors.New("deployment id must not contain whitespace")
	ErrInvalidDeploymentName      = errors.New("deployment name must not be blank")
	ErrDeploymentNotFound         = errors.New("deployment not found")
	ErrDeploymentInactive         = errors.New("deployment is deactivated")
	ErrDeploymentLimitReached     = errors.New("license deployment limit reached")
	ErrActivationConflict         = errors.New("deployment activation conflicted with another concurrent change")
)

// License is Leamout-owned commercial authority for self-hosted installations.
// The signed artifact format is intentionally separate from this persistence model.
type License struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	SubscriptionID *uuid.UUID `json:"subscription_id,omitempty"`
	Status         Status     `json:"status"`
	MaxDeployments int32      `json:"max_deployments"`
	SigningKeyID   *string    `json:"signing_key_id,omitempty"`
	IssuedAt       time.Time  `json:"issued_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Deployment is one activated self-hosted installation under a license.
type Deployment struct {
	ID             uuid.UUID        `json:"id"`
	OrganizationID uuid.UUID        `json:"organization_id"`
	LicenseID      uuid.UUID        `json:"license_id"`
	DeploymentID   string           `json:"deployment_id"`
	Name           *string          `json:"name,omitempty"`
	Status         DeploymentStatus `json:"status"`
	ActivatedAt    time.Time        `json:"activated_at"`
	LastSeenAt     *time.Time       `json:"last_seen_at,omitempty"`
	DeactivatedAt  *time.Time       `json:"deactivated_at,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// CreateInput contains trusted licensing-authority metadata. Subscription identity
// and deployment limits are resolved from current commercial state by the service.
type CreateInput struct {
	SigningKeyID *string    `json:"signing_key_id,omitempty"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// ActivateDeploymentInput identifies a stable installation requesting a license slot.
type ActivateDeploymentInput struct {
	DeploymentID string  `json:"deployment_id"`
	Name         *string `json:"name,omitempty"`
}
