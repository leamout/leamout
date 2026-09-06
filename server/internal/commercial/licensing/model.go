package licensing

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
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

const (
	MaxDeploymentsEntitlement = "max.deployments"
	ManagedVoiceEntitlement    = "voice.managed.enabled"

	ManagedCarrierPurpose      = "managed_carrier"
	ManagedCarrierTransitScope = "managed_carrier:transit"
)

var (
	ErrLicenseNotFound             = apperror.NewNotFound("license not found")
	ErrLicenseUnavailable          = apperror.NewConflict("license is unavailable")
	ErrCommercialStateUnavailable  = apperror.NewPaymentRequired("commercial state does not allow license creation")
	ErrManagedCarrierUnavailable   = apperror.NewPaymentRequired("managed carrier is not enabled for this organization")
	ErrOrganizationIDRequired      = apperror.NewBadRequest("organization id is required")
	ErrLicenseIDRequired           = apperror.NewBadRequest("license id is required")
	ErrSigningKeyRequired          = errors.New("signing key id is required before license activation")
	ErrInvalidSigningKey           = errors.New("signing key id must not contain whitespace")
	ErrSigningKeyUnavailable       = errors.New("license signing key is unavailable")
	ErrUnsupportedLicenseVersion   = errors.New("unsupported signed license version")
	ErrUnsupportedAlgorithm        = errors.New("unsupported license signature algorithm")
	ErrMalformedArtifact           = errors.New("malformed signed license artifact")
	ErrInvalidSignature            = errors.New("invalid signed license signature")
	ErrArtifactExpired             = errors.New("signed license artifact has expired")
	ErrArtifactNotYetValid         = errors.New("signed license artifact is not yet valid")
	ErrDeploymentMismatch          = errors.New("signed license is bound to another deployment")
	ErrInvalidStatus               = errors.New("invalid license status")
	ErrInvalidTransition           = errors.New("invalid license status transition")
	ErrInvalidDeploymentLimit      = apperror.NewConflict("max deployments must be greater than zero")
	ErrInvalidExpiration           = apperror.NewBadRequest("license expiration must be after issuance")
	ErrDeploymentIDRequired        = apperror.NewBadRequest("deployment_id is required")
	ErrInvalidDeploymentID         = apperror.NewBadRequest("deployment_id must not contain whitespace")
	ErrInvalidDeploymentName       = apperror.NewBadRequest("deployment name must not be blank")
	ErrDeploymentNotFound          = apperror.NewNotFound("deployment not found")
	ErrDeploymentInactive          = apperror.NewConflict("deployment is deactivated")
	ErrDeploymentLimitReached      = apperror.NewConflict("license deployment limit reached")
	ErrActivationConflict          = apperror.NewConflict("deployment activation conflicted with another concurrent change")
	ErrInvalidDeploymentCredential = apperror.NewUnauthorized("invalid deployment credential")
	ErrDeploymentCredentialScope   = apperror.NewForbidden("deployment credential does not authorize this operation")
)

// License is Leamout-owned commercial authority for self-hosted installations.
// The signed artifact format is intentionally separate from this persistence model.
type License struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	SubscriptionID *uuid.UUID
	Status         Status
	MaxDeployments int32
	SigningKeyID   *string
	IssuedAt       time.Time
	ExpiresAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Deployment is one activated self-hosted installation under a license.
type Deployment struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	LicenseID      uuid.UUID
	DeploymentID   string
	Name           *string
	Status         DeploymentStatus
	ActivatedAt    time.Time
	LastSeenAt     *time.Time
	DeactivatedAt  *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// DeploymentCredential is a cloud-authoritative machine identity for one deployment.
// The usable token is returned once at enrollment time and is never persisted.
type DeploymentCredential struct {
	ID              uuid.UUID
	DeploymentRowID uuid.UUID
	Purpose         string
	TokenPrefix     string
	Scopes          []string
	ExpiresAt       *time.Time
	LastUsedAt      *time.Time
	DisabledAt      *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type deploymentCredentialAuthentication struct {
	CredentialID     uuid.UUID
	OrganizationID   uuid.UUID
	LicenseID        uuid.UUID
	DeploymentID     string
	DeploymentStatus string
	LicenseStatus    string
	LicenseExpiresAt *time.Time
	Purpose          string
	Scopes           []string
}

// DeploymentIdentity is the trusted identity returned after authenticating a
// deployment credential at the Leamout control-plane boundary.
type DeploymentIdentity struct {
	CredentialID   uuid.UUID
	OrganizationID uuid.UUID
	LicenseID      uuid.UUID
	DeploymentID   string
	Purpose        string
	Scopes         []string
}

// ManagedCarrierEnrollment contains the one-time deployment credential issued for
// self-hosted access to Leamout-managed carrier services.
type ManagedCarrierEnrollment struct {
	DeploymentID string
	Token        string
	TokenPrefix  string
	Scopes       []string
	ExpiresAt    *time.Time
}

// CreateInput contains trusted licensing-authority metadata. Subscription identity
// and deployment limits are resolved from current commercial state by the service.
type CreateInput struct {
	SigningKeyID *string
	ExpiresAt    *time.Time
}

type entitlementSnapshot struct {
	Features map[string]bool
	Limits   map[string]int64
}

// ActivateDeploymentInput identifies a stable installation requesting a license slot.
type ActivateDeploymentInput struct {
	DeploymentID string  `json:"deployment_id"`
	Name         *string `json:"name,omitempty"`
}

type licenseResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	SubscriptionID *uuid.UUID `json:"subscription_id,omitempty"`
	Status         Status     `json:"status"`
	MaxDeployments int32      `json:"max_deployments"`
	IssuedAt       time.Time  `json:"issued_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func newLicenseResponse(license License) licenseResponse {
	return licenseResponse{
		ID: license.ID, OrganizationID: license.OrganizationID, SubscriptionID: license.SubscriptionID,
		Status: license.Status, MaxDeployments: license.MaxDeployments, IssuedAt: license.IssuedAt,
		ExpiresAt: license.ExpiresAt, CreatedAt: license.CreatedAt, UpdatedAt: license.UpdatedAt,
	}
}

type deploymentResponse struct {
	ID            uuid.UUID        `json:"id"`
	LicenseID     uuid.UUID        `json:"license_id"`
	DeploymentID  string           `json:"deployment_id"`
	Name          *string          `json:"name,omitempty"`
	Status        DeploymentStatus `json:"status"`
	ActivatedAt   time.Time        `json:"activated_at"`
	LastSeenAt    *time.Time       `json:"last_seen_at,omitempty"`
	DeactivatedAt *time.Time       `json:"deactivated_at,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

func newDeploymentResponse(deployment Deployment) deploymentResponse {
	return deploymentResponse{
		ID: deployment.ID, LicenseID: deployment.LicenseID, DeploymentID: deployment.DeploymentID,
		Name: deployment.Name, Status: deployment.Status, ActivatedAt: deployment.ActivatedAt,
		LastSeenAt: deployment.LastSeenAt, DeactivatedAt: deployment.DeactivatedAt,
		CreatedAt: deployment.CreatedAt, UpdatedAt: deployment.UpdatedAt,
	}
}

type managedCarrierEnrollmentResponse struct {
	DeploymentID string     `json:"deployment_id"`
	Token        string     `json:"token"`
	TokenPrefix  string     `json:"token_prefix"`
	Scopes       []string   `json:"scopes"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

func newManagedCarrierEnrollmentResponse(enrollment ManagedCarrierEnrollment) managedCarrierEnrollmentResponse {
	return managedCarrierEnrollmentResponse(enrollment)
}
