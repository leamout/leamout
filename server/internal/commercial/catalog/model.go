package catalog

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// BillingInterval identifies the recurring cadence attached to a catalog price.
type BillingInterval string

const (
	BillingIntervalMonth BillingInterval = "month"
	BillingIntervalYear  BillingInterval = "year"
)

var (
	ErrProductNotFound        = errors.New("catalog product not found")
	ErrPlanNotFound           = errors.New("catalog plan not found")
	ErrPriceNotFound          = errors.New("catalog price not found")
	ErrCodeRequired           = errors.New("catalog code is required")
	ErrInvalidCode            = errors.New("catalog code must not contain whitespace")
	ErrIDRequired             = errors.New("catalog id is required")
	ErrInvalidBillingInterval = errors.New("invalid catalog billing interval")
)

// Product is a commercial product family that groups reusable plans.
type Product struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Plan is a reusable commercial offer within a product.
type Plan struct {
	ID          uuid.UUID `json:"id"`
	ProductID   uuid.UUID `json:"product_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Price is an immutable set of recurring commercial terms for a plan.
// Active/effective bounds control acquisition availability without rewriting
// historical subscriptions that already reference this price.
type Price struct {
	ID              uuid.UUID       `json:"id"`
	PlanID          uuid.UUID       `json:"plan_id"`
	Currency        string          `json:"currency"`
	AmountMinor     int64           `json:"amount_minor"`
	BillingInterval BillingInterval `json:"billing_interval"`
	Active          bool            `json:"active"`
	EffectiveFrom   time.Time       `json:"effective_from"`
	EffectiveUntil  *time.Time      `json:"effective_until,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
}

func (p Price) EffectiveAt(at time.Time) bool {
	if !p.Active || at.Before(p.EffectiveFrom) {
		return false
	}
	return p.EffectiveUntil == nil || at.Before(*p.EffectiveUntil)
}
