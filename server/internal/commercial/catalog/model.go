package catalog

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

type PricingType string
type BillingInterval string

const (
	PricingTypeOneTime   PricingType = "one_time"
	PricingTypeRecurring PricingType = "recurring"
	PricingTypeMetered   PricingType = "metered"

	BillingIntervalMonth BillingInterval = "month"
	BillingIntervalYear  BillingInterval = "year"
)

var (
	ErrProductNotFound = apperror.NewNotFound("catalog product not found")
	ErrPlanNotFound    = apperror.NewNotFound("catalog plan not found")
	ErrPriceNotFound   = apperror.NewNotFound("catalog price not found")
	ErrMeterNotFound   = apperror.NewNotFound("catalog meter not found")
	ErrCodeRequired    = errors.New("catalog code is required")
	ErrInvalidCode     = errors.New("catalog code must not contain whitespace")
	ErrIDRequired      = errors.New("catalog id is required")
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

// Meter identifies a quantity that can be referenced by metered prices.
type Meter struct {
	ID        uuid.UUID
	Key       string
	Name      string
	Unit      string
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Price is an immutable set of customer-facing commercial terms for a plan.
// A price is either one-time, recurring, or metered. Metered prices reference
// the meter they rate and may use dimensions for routing-specific variants.
type Price struct {
	ID               uuid.UUID
	PlanID           uuid.UUID
	MeterID          *uuid.UUID
	PricingType      PricingType
	Currency         string
	AmountMinor      *int64
	BillingInterval  *BillingInterval
	UnitAmountMicros *int64
	UnitSize         *int64
	Dimensions       json.RawMessage
	Active           bool
	EffectiveFrom    time.Time
	EffectiveUntil   *time.Time
	CreatedAt        time.Time
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
	ID               string           `json:"id"`
	PlanID           string           `json:"plan_id"`
	MeterID          *uuid.UUID       `json:"meter_id,omitempty"`
	PricingType      PricingType      `json:"pricing_type"`
	Currency         string           `json:"currency"`
	AmountMinor      *int64           `json:"amount_minor,omitempty"`
	BillingInterval  *BillingInterval `json:"billing_interval,omitempty"`
	UnitAmountMicros *int64           `json:"unit_amount_micros,omitempty"`
	UnitSize         *int64           `json:"unit_size,omitempty"`
	Dimensions       json.RawMessage  `json:"dimensions,omitempty"`
	EffectiveFrom    time.Time        `json:"effective_from"`
	EffectiveUntil   *time.Time       `json:"effective_until,omitempty"`
}

func newPriceResponse(price Price) priceResponse {
	return priceResponse{
		ID:               price.ID.String(),
		PlanID:           price.PlanID.String(),
		MeterID:          price.MeterID,
		PricingType:      price.PricingType,
		Currency:         price.Currency,
		AmountMinor:      price.AmountMinor,
		BillingInterval:  price.BillingInterval,
		UnitAmountMicros: price.UnitAmountMicros,
		UnitSize:         price.UnitSize,
		Dimensions:       price.Dimensions,
		EffectiveFrom:    price.EffectiveFrom,
		EffectiveUntil:   price.EffectiveUntil,
	}
}

func (p Price) EffectiveAt(at time.Time) bool {
	if !p.Active || at.Before(p.EffectiveFrom) {
		return false
	}
	return p.EffectiveUntil == nil || at.Before(*p.EffectiveUntil)
}
