package polar

import (
	"encoding/json"
	"time"
)

const (
	DeveloperProductID = "58ff657a-74d9-4342-89c4-2a3b784c6965"
	ProProductID       = "da5b3818-e92e-4b09-afad-ae61895f05db"
	ScaleProductID     = "a3023020-9822-4835-9509-1635c7d35f63"
	APIVersion         = "2026-10"
)

type CheckoutParams struct {
	ProductID      string
	OrganizationID string
	PlanCode       string
	SuccessURL     string
	ReturnURL      string
}

type CheckoutSession struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type PortalParams struct {
	CustomerID string
	ReturnURL  string
}

type PortalSession struct {
	URL string `json:"customer_portal_url"`
}

type Customer struct {
	ID         string `json:"id"`
	ExternalID string `json:"external_id"`
}

type Subscription struct {
	ID                 string         `json:"id"`
	CustomerID         string         `json:"customer_id"`
	ProductID          string         `json:"product_id"`
	Customer           Customer       `json:"customer"`
	Status             string         `json:"status"`
	Metadata           map[string]any `json:"metadata"`
	CurrentPeriodStart time.Time      `json:"current_period_start"`
	CurrentPeriodEnd   time.Time      `json:"current_period_end"`
	TrialEnd           *time.Time     `json:"trial_end"`
	CancelAtPeriodEnd  bool           `json:"cancel_at_period_end"`
	CanceledAt         *time.Time     `json:"canceled_at"`
}

type Event struct {
	Type       string          `json:"type"`
	Timestamp  time.Time       `json:"timestamp"`
	APIVersion string          `json:"api_version"`
	Data       json.RawMessage `json:"data"`
}
