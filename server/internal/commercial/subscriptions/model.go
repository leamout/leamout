package subscriptions

import (
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

// Status describes the commercial lifecycle of a subscription.
type Status string

const (
	StatusPending   Status = "pending"
	StatusActive    Status = "active"
	StatusPastDue   Status = "past_due"
	StatusCancelled Status = "cancelled"
	StatusExpired   Status = "expired"
)

var (
	ErrSubscriptionNotFound      = apperror.NewNotFound("subscription not found")
	ErrOrganizationUnavailable   = apperror.NewConflict("organization is unavailable for subscription changes")
	ErrCurrentSubscriptionExists = apperror.NewConflict("organization already has a current subscription")
	ErrPriceUnavailable          = apperror.NewBadRequest("subscription price is unavailable")
	ErrInvalidStatus             = apperror.NewBadRequest("invalid subscription status")
	ErrInvalidInitialStatus      = apperror.NewBadRequest("subscription must start pending or active")
	ErrInvalidTransition         = apperror.NewConflict("invalid subscription status transition")
	ErrInvalidPeriod             = apperror.NewBadRequest("invalid subscription period")
	ErrOrganizationIDRequired    = apperror.NewBadRequest("organization id is required")
	ErrSubscriptionIDRequired    = apperror.NewBadRequest("subscription id is required")
	ErrPriceIDRequired           = apperror.NewBadRequest("price_id is required")
	ErrNoChanges                 = apperror.NewBadRequest("at least one subscription field is required")
	ErrTerminalSubscription      = apperror.NewConflict("terminal subscription cannot change commercial terms")
)

// Subscription binds an organization to an acquired recurring Leamout price.
// Payment providers settle checkout orders; they do not own subscription state.
type Subscription struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	PlanID         uuid.UUID
	PriceID        uuid.UUID
	Status         Status
	StartsAt       time.Time
	RenewsAt       *time.Time
	EndsAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateInput struct {
	PriceID  uuid.UUID
	Status   *Status
	StartsAt *time.Time
	RenewsAt *time.Time
	EndsAt   *time.Time
}

type PeriodUpdate struct {
	RenewsAt *time.Time
	EndsAt   *time.Time
}

type UpdateRequest struct {
	PriceID uuid.UUID `json:"price_id"`
}

type subscriptionResponse struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	PlanID         uuid.UUID  `json:"plan_id"`
	PriceID        uuid.UUID  `json:"price_id"`
	Status         Status     `json:"status"`
	StartsAt       time.Time  `json:"starts_at"`
	RenewsAt       *time.Time `json:"renews_at,omitempty"`
	EndsAt         *time.Time `json:"ends_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func newSubscriptionResponse(subscription Subscription) subscriptionResponse {
	return subscriptionResponse{
		ID:             subscription.ID,
		OrganizationID: subscription.OrganizationID,
		PlanID:         subscription.PlanID,
		PriceID:        subscription.PriceID,
		Status:         subscription.Status,
		StartsAt:       subscription.StartsAt,
		RenewsAt:       subscription.RenewsAt,
		EndsAt:         subscription.EndsAt,
		CreatedAt:      subscription.CreatedAt,
		UpdatedAt:      subscription.UpdatedAt,
	}
}
