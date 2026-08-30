package subscriptions

import "time"

// Status describes the commercial lifecycle of a subscription.
type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusPastDue   Status = "past_due"
	StatusCancelled Status = "cancelled"
	StatusExpired   Status = "expired"
)

// Subscription binds a customer to a commercial plan for a period of time.
type Subscription struct {
	ID             string
	OrganizationID string
	PlanID         string
	Status         Status
	StartsAt       time.Time
	RenewsAt       *time.Time
	EndsAt         *time.Time
}
