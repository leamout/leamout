package licensing

import (
	"time"

	"github.com/google/uuid"
)

// Status describes whether a commercial license can be activated or renewed.
type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusExpired   Status = "expired"
	StatusRevoked   Status = "revoked"
)

// DocumentVersion is the compatibility version of the signed claim payload.
const DocumentVersion = 1

type IssueRequest struct {
	OrganizationID uuid.UUID
	LicenseID      uuid.UUID
	At             time.Time
}

type Claims struct {
	Version        int              `json:"version"`
	OrganizationID uuid.UUID        `json:"organization_id"`
	LicenseID      uuid.UUID        `json:"license_id"`
	IssuedAt       time.Time        `json:"issued_at"`
	ExpiresAt      *time.Time       `json:"expires_at,omitempty"`
	MaxDeployments int32            `json:"max_deployments"`
	Features       map[string]bool  `json:"features"`
	Limits         map[string]int64 `json:"limits"`
}

type Document struct {
	Algorithm string `json:"algorithm"`
	KeyID     string `json:"key_id"`
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
}
