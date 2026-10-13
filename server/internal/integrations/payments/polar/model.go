package polar

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	DeveloperProductID = "58ff657a-74d9-4342-89c4-2a3b784c6965"
	ProProductID       = "da5b3818-e92e-4b09-afad-ae61895f05db"
	ScaleProductID     = "a3023020-9822-4835-9509-1635c7d35f63"
	APIVersion         = "2026-10"
)

var ErrNotFound = errors.New("polar resource not found")

type APIError struct {
	StatusCode int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("polar request failed with status %d", e.StatusCode)
}

func (e *APIError) Is(target error) bool {
	return target == ErrNotFound && e.StatusCode == http.StatusNotFound
}

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

type CustomerState struct {
	ID                  string                      `json:"id"`
	ExternalID          string                      `json:"external_id"`
	ActiveSubscriptions []CustomerStateSubscription `json:"active_subscriptions"`
}

// CustomerStateSubscription contains active or trialing subscriptions. Polar's
// state response omits past-due subscriptions and does not embed the customer.
type CustomerStateSubscription struct {
	ID                 string     `json:"id"`
	ProductID          string     `json:"product_id"`
	Status             string     `json:"status"`
	CurrentPeriodStart time.Time  `json:"current_period_start"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end"`
	TrialEnd           *time.Time `json:"trial_end"`
	CancelAtPeriodEnd  bool       `json:"cancel_at_period_end"`
	CanceledAt         *time.Time `json:"canceled_at"`
}

type Subscription struct {
	ID                 string         `json:"id"`
	ModifiedAt         *time.Time     `json:"modified_at"`
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
	PastDueAt          *time.Time     `json:"past_due_at"`
}

type Event struct {
	Type       string          `json:"type"`
	Timestamp  time.Time       `json:"timestamp"`
	APIVersion string          `json:"api_version"`
	Data       json.RawMessage `json:"data"`
}
