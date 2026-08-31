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
	ID          uuid.UUID
	Code        string
	Name        string
	Description *string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Plan is a reusable commercial offer within a product.
type Plan struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	Code        string
	Name        string
	Description *string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Price is an immutable set of recurring commercial terms for a plan.
// Active/effective bounds control acquisition availability without rewriting
// historical subscriptions that already reference this price.
type Price struct {
	ID              uuid.UUID
	PlanID          uuid.UUID
	Currency        string
	AmountMinor     int64
	BillingInterval BillingInterval
	Active          bool
	EffectiveFrom   time.Time
	EffectiveUntil  *time.Time
	CreatedAt       time.Time
}

func (p Price) EffectiveAt(at time.Time) bool {
	if !p.Active || at.Before(p.EffectiveFrom) {
		return false
	}
	return p.EffectiveUntil == nil || at.Before(*p.EffectiveUntil)
}
