// Package email owns transactional email rendering and durable delivery.
package email

import (
	"time"

	"github.com/google/uuid"
)

type Data struct {
	ChangedAt    string    `json:"changed_at,omitempty"`
	RecoveryURL  string    `json:"recovery_url,omitempty"`
	Code         string    `json:"code,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Organization string    `json:"organization,omitempty"`
	Inviter      string    `json:"inviter,omitempty"`
	Role         string    `json:"role,omitempty"`
	AcceptURL    string    `json:"accept_url,omitempty"`
}
type Request struct {
	To              string
	Template        string
	Data            Data
	CancellationKey *string
}
type Delivery struct {
	ID                    uuid.UUID
	To, Template, Payload string
	Attempts              int
	ExpiresAt             time.Time
}
