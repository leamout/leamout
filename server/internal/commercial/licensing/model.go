package licensing

import "time"

// Status describes whether a commercial license can be activated or renewed.
type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
	StatusExpired   Status = "expired"
	StatusRevoked   Status = "revoked"
)

// License is the commercial authority from which deployment entitlements are issued.
type License struct {
	ID             string
	CustomerID     string
	SubscriptionID string
	Status         Status
	MaxDeployments int
	IssuedAt       time.Time
	ExpiresAt      *time.Time
}
