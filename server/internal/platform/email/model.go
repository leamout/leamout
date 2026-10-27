// Package email owns transactional email rendering and durable delivery.
package email

import (
	"time"

	"github.com/google/uuid"
)

type Data struct {
	ExpiresAt        time.Time `json:"expires_at"`
	Organization     string    `json:"organization,omitempty"`
	Inviter          string    `json:"inviter,omitempty"`
	Role             string    `json:"role,omitempty"`
	AcceptURL        string    `json:"accept_url,omitempty"`
	BillingURL       string    `json:"billing_url,omitempty"`
	UsedMinutes      int64     `json:"used_minutes,omitempty"`
	AllowanceMinutes int64     `json:"allowance_minutes,omitempty"`
	UsagePercent     int64     `json:"usage_percent,omitempty"`
	PeriodEnd        time.Time `json:"period_end,omitempty"`
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
