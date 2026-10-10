package plans

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusActive   = "active"
	StatusInactive = "inactive"

	IntervalMonth = "month"
	IntervalYear  = "year"
)

type Plan struct {
	ID              uuid.UUID
	Code            string
	Name            string
	Description     *string
	Currency        string
	AmountMinor     int64
	BillingInterval string
	Entitlements    map[string]bool
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type UpsertRequest struct {
	Code            string
	Name            string
	Description     *string
	Currency        string
	AmountMinor     int64
	BillingInterval string
	Entitlements    map[string]bool
	Status          string
}

type Response struct {
	ID              uuid.UUID       `json:"id"`
	Code            string          `json:"code"`
	Name            string          `json:"name"`
	Description     *string         `json:"description,omitempty"`
	Currency        string          `json:"currency"`
	AmountMinor     int64           `json:"amount_minor"`
	BillingInterval string          `json:"billing_interval"`
	Entitlements    map[string]bool `json:"entitlements"`
	Status          string          `json:"status"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func response(value Plan) Response {
	return Response(value)
}
