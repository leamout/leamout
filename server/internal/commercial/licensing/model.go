package licensing

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusExpired   Status = "expired"
	StatusRevoked   Status = "revoked"
)

type DeploymentStatus string

const (
	DeploymentStatusActive      DeploymentStatus = "active"
	DeploymentStatusDeactivated DeploymentStatus = "deactivated"
)

var (
	ErrLicenseNotFound              = apperror.NewNotFound("license not found")
	ErrLicenseUnavailable           = apperror.NewConflict("license is unavailable")
	ErrOrganizationIDRequired       = apperror.NewBadRequest("organization id is required")
	ErrLicenseIDRequired            = apperror.NewBadRequest("license id is required")
	ErrSigningKeyRequired           = errors.New("signing key id is required before license activation")
	ErrInvalidSigningKey            = errors.New("signing key id must not contain whitespace")
	ErrSigningKeyUnavailable        = errors.New("license signing key is unavailable")
	ErrUnsupportedLicenseVersion    = errors.New("unsupported signed license version")
	ErrUnsupportedAlgorithm         = errors.New("unsupported license signature algorithm")
	ErrMalformedArtifact            = errors.New("malformed signed license artifact")
	ErrInvalidSignature             = errors.New("invalid signed license signature")
	ErrArtifactExpired              = errors.New("signed license artifact has expired")
	ErrArtifactNotYetValid          = errors.New("signed license artifact is not yet valid")
	ErrDeploymentMismatch           = errors.New("signed license is bound to another deployment")
	ErrDeploymentKeyMismatch        = errors.New("signed license is bound to another deployment key")
	ErrInvalidStatus                = errors.New("invalid license status")
	ErrInvalidTransition            = errors.New("invalid license status transition")
	ErrInvalidExpiration            = apperror.NewBadRequest("license expiration must be after issuance")
	ErrDeploymentIDRequired         = apperror.NewBadRequest("deployment_id is required")
	ErrInvalidDeploymentID          = apperror.NewBadRequest("deployment_id must not contain whitespace")
	ErrDeploymentPublicKeyRequired  = apperror.NewBadRequest("public_key is required")
	ErrInvalidDeploymentPublicKey   = apperror.NewBadRequest("public_key must be a base64url Ed25519 public key")
	ErrInvalidDeploymentName        = apperror.NewBadRequest("deployment name must not be blank")
	ErrDeploymentNotFound           = apperror.NewNotFound("deployment not found")
	ErrDeploymentInactive           = apperror.NewConflict("deployment is deactivated")
	ErrLicenseAlreadyBound          = apperror.NewConflict("license is already bound to another deployment")
	ErrActivationConflict           = apperror.NewConflict("deployment activation conflicted with another concurrent change")
)

type License struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	Status         Status
	SigningKeyID   *string
	IssuedAt       time.Time
	ExpiresAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Deployment struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	LicenseID      uuid.UUID
	DeploymentID   string
	PublicKey      string
	Name           *string
	Status         DeploymentStatus
	ActivatedAt    time.Time
	LastSeenAt     *time.Time
	DeactivatedAt  *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CreateInput is explicit self-hosted licensing authority. It is independent
// of Cloud PAYG and authorizes one deployment.
type CreateInput struct {
	SigningKeyID *string
	ExpiresAt    *time.Time
}

type ActivateDeploymentInput struct {
	DeploymentID string  `json:"deployment_id"`
	PublicKey    string  `json:"public_key"`
	Name         *string `json:"name,omitempty"`
}

type licenseResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	Status         Status     `json:"status"`
	IssuedAt       time.Time  `json:"issued_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func newLicenseResponse(license License) licenseResponse {
	return licenseResponse{
		ID: license.ID, OrganizationID: license.OrganizationID, Status: license.Status,
		IssuedAt: license.IssuedAt, ExpiresAt: license.ExpiresAt,
		CreatedAt: license.CreatedAt, UpdatedAt: license.UpdatedAt,
	}
}

type deploymentResponse struct {
	ID            uuid.UUID        `json:"id"`
	LicenseID     uuid.UUID        `json:"license_id"`
	DeploymentID  string           `json:"deployment_id"`
	PublicKey     string           `json:"public_key"`
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
		PublicKey: deployment.PublicKey, Name: deployment.Name, Status: deployment.Status, ActivatedAt: deployment.ActivatedAt,
		LastSeenAt: deployment.LastSeenAt, DeactivatedAt: deployment.DeactivatedAt,
		CreatedAt: deployment.CreatedAt, UpdatedAt: deployment.UpdatedAt,
	}
}
