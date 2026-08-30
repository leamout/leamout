package deployments

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusActive      Status = "active"
	StatusDeactivated Status = "deactivated"
)

type Deployment struct {
	ID            uuid.UUID
	LicenseID     uuid.UUID
	DeploymentID  string
	Name          *string
	Status        Status
	ActivatedAt   time.Time
	LastSeenAt    *time.Time
	DeactivatedAt *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type ActivateRequest struct {
	OrganizationID uuid.UUID
	LicenseID      uuid.UUID
	DeploymentID   string
	Name           *string
	At             time.Time
}
