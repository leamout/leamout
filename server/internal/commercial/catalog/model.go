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

type productResponse struct {
	ID          string  `json:"id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func newProductResponse(product Product) productResponse {
	return productResponse{
		ID:          product.ID.String(),
		Code:        product.Code,
		Name:        product.Name,
		Description: product.Description,
	}
}

type planResponse struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	Code        string  `json:"code"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func newPlanResponse(plan Plan) planResponse {
	return planResponse{
		ID:          plan.ID.String(),
		ProductID:   plan.ProductID.String(),
		Code:        plan.Code,
		Name:        plan.Name,
		Description: plan.Description,
	}
}

type priceResponse struct {
	ID              string          `json:"id"`
	PlanID          string          `json:"plan_id"`
	Currency        string          `json:"currency"`
	AmountMinor     int64           `json:"amount_minor"`
	BillingInterval BillingInterval `json:"billing_interval"`
	EffectiveFrom   time.Time       `json:"effective_from"`
	EffectiveUntil  *time.Time      `json:"effective_until,omitempty"`
}

func newPriceResponse(price Price) priceResponse {
	return priceResponse{
		ID:              price.ID.String(),
		PlanID:          price.PlanID.String(),
		Currency:        price.Currency,
		AmountMinor:     price.AmountMinor,
		BillingInterval: price.BillingInterval,
		EffectiveFrom:   price.EffectiveFrom,
		EffectiveUntil:  price.EffectiveUntil,
	}
}

func (p Price) EffectiveAt(at time.Time) bool {
	if !p.Active || at.Before(p.EffectiveFrom) {
		return false
	}
	return p.EffectiveUntil == nil || at.Before(*p.EffectiveUntil)
}
