package subscriptions

import (
	"errors"
	"time"

	"github.com/google/uuid"
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
	ErrSubscriptionNotFound    = errors.New("subscription not found")
	ErrOrganizationUnavailable = errors.New("organization is unavailable for subscription changes")
	ErrPriceUnavailable        = errors.New("subscription price is unavailable")
	ErrInvalidStatus           = errors.New("invalid subscription status")
	ErrInvalidInitialStatus    = errors.New("subscription must start pending or active")
	ErrInvalidTransition       = errors.New("invalid subscription status transition")
	ErrInvalidPeriod           = errors.New("invalid subscription period")
	ErrProviderConflict        = errors.New("provider subscription identifier already exists")
	ErrProviderRequired        = errors.New("billing provider is required")
	ErrProviderIDRequired      = errors.New("provider subscription id is required")
	ErrInvalidProvider         = errors.New("billing provider must not contain whitespace")
	ErrOrganizationIDRequired  = errors.New("organization id is required")
	ErrSubscriptionIDRequired  = errors.New("subscription id is required")
	ErrPriceIDRequired         = errors.New("price id is required")
	ErrNoChanges               = errors.New("at least one subscription field is required")
	ErrTerminalSubscription    = errors.New("terminal subscription cannot change commercial terms")
)

// Subscription binds an organization to a commercial plan and acquired price for a period of time.
// PriceID can be nil only for legacy rows created before price-backed subscriptions were introduced.
type Subscription struct {
	ID                     uuid.UUID
	OrganizationID         uuid.UUID
	PlanID                 uuid.UUID
	PriceID                *uuid.UUID
	Status                 Status
	StartsAt               time.Time
	RenewsAt               *time.Time
	EndsAt                 *time.Time
	BillingProvider        *string
	ProviderSubscriptionID *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// ProviderReference identifies the matching subscription at an external billing provider.
// It is reconciliation metadata; the provider is not the source of truth for subscription state.
type ProviderReference struct {
	Provider       string
	SubscriptionID string
}

// CreateInput describes a new organization subscription. The selected price determines the plan.
type CreateInput struct {
	PriceID  uuid.UUID
	Status   *Status
	StartsAt *time.Time
	RenewsAt *time.Time
	EndsAt   *time.Time
	Provider *ProviderReference
}

// PeriodUpdate changes future renewal/end timestamps without changing subscription identity.
type PeriodUpdate struct {
	RenewsAt *time.Time
	EndsAt   *time.Time
}
