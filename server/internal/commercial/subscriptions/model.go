package subscriptions

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusTrialing   = "trialing"
	StatusActive     = "active"
	StatusPastDue    = "past_due"
	StatusCanceled   = "canceled"
	StatusIncomplete = "incomplete"
)

type Subscription struct {
	ID                 uuid.UUID
	OrganizationID     uuid.UUID
	PlanID             uuid.UUID
	Status             string
	CurrentPeriodStart *time.Time
	CurrentPeriodEnd   *time.Time
	TrialEndsAt        *time.Time
	CancelAtPeriodEnd  bool
	CanceledAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type SyncRequest struct {
	PlanID             uuid.UUID
	Status             string
	CurrentPeriodStart *time.Time
	CurrentPeriodEnd   *time.Time
	TrialEndsAt        *time.Time
	CancelAtPeriodEnd  bool
	CanceledAt         *time.Time
}

type Response struct {
	ID                 uuid.UUID  `json:"id"`
	OrganizationID     uuid.UUID  `json:"organization_id"`
	PlanID             uuid.UUID  `json:"plan_id"`
	Status             string     `json:"status"`
	CurrentPeriodStart *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd   *time.Time `json:"current_period_end,omitempty"`
	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end"`
	CanceledAt         *time.Time `json:"canceled_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func response(value Subscription) Response {
	return Response(value)
}
