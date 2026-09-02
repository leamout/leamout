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
	ErrSubscriptionNotFound      = errors.New("subscription not found")
	ErrOrganizationUnavailable   = errors.New("organization is unavailable for subscription changes")
	ErrCurrentSubscriptionExists = errors.New("organization already has a current subscription")
	ErrPriceUnavailable          = errors.New("subscription price is unavailable")
	ErrInvalidStatus             = errors.New("invalid subscription status")
	ErrInvalidInitialStatus      = errors.New("subscription must start pending or active")
	ErrInvalidTransition         = errors.New("invalid subscription status transition")
	ErrInvalidPeriod             = errors.New("invalid subscription period")
	ErrProviderConflict          = errors.New("provider subscription identifier already exists")
	ErrProviderRequired          = errors.New("billing provider is required")
	ErrProviderIDRequired        = errors.New("provider subscription id is required")
	ErrInvalidProvider           = errors.New("billing provider must not contain whitespace")
	ErrOrganizationIDRequired    = errors.New("organization id is required")
	ErrSubscriptionIDRequired    = errors.New("subscription id is required")
	ErrPriceIDRequired           = errors.New("price id is required")
	ErrNoChanges                 = errors.New("at least one subscription field is required")
	ErrTerminalSubscription      = errors.New("terminal subscription cannot change commercial terms")
)

// Subscription binds an organization to a commercial plan and acquired price for a period of time.
// PriceID can be nil only for legacy rows created before price-backed subscriptions were introduced.
type Subscription struct {
	ID                     uuid.UUID  `json:"id"`
	OrganizationID         uuid.UUID  `json:"organization_id"`
	PlanID                 uuid.UUID  `json:"plan_id"`
	PriceID                *uuid.UUID `json:"price_id,omitempty"`
	Status                 Status     `json:"status"`
	StartsAt               time.Time  `json:"starts_at"`
	RenewsAt               *time.Time `json:"renews_at,omitempty"`
	EndsAt                 *time.Time `json:"ends_at,omitempty"`
	BillingProvider        *string    `json:"billing_provider,omitempty"`
	ProviderSubscriptionID *string    `json:"provider_subscription_id,omitempty"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

// ProviderReference identifies the matching subscription at an external billing provider.
// It is reconciliation metadata; the provider is not the source of truth for subscription state.
type ProviderReference struct {
	Provider       string `json:"provider"`
	SubscriptionID string `json:"subscription_id"`
}

// CreateInput describes a new organization subscription. The selected price determines the plan.
type CreateInput struct {
	PriceID  uuid.UUID          `json:"price_id"`
	Status   *Status            `json:"status,omitempty"`
	StartsAt *time.Time         `json:"starts_at,omitempty"`
	RenewsAt *time.Time         `json:"renews_at,omitempty"`
	EndsAt   *time.Time         `json:"ends_at,omitempty"`
	Provider *ProviderReference `json:"provider,omitempty"`
}

// PeriodUpdate changes future renewal/end timestamps without changing subscription identity.
type PeriodUpdate struct {
	RenewsAt *time.Time `json:"renews_at"`
	EndsAt   *time.Time `json:"ends_at"`
}
