package billing

import (
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/commercial/limits"
)

type CheckoutRequest struct {
	PlanCode   string `json:"plan_code"`
	SuccessURL string `json:"success_url"`
	CancelURL  string `json:"cancel_url"`
}

type PortalRequest struct {
	ReturnURL string `json:"return_url"`
}

type SessionResponse struct {
	URL string `json:"url"`
}

type ReconciliationTarget struct {
	OrganizationID uuid.UUID
	SubscriptionID string
	CustomerID     string
}

type PlanSummary struct {
	Code            string `json:"code"`
	Name            string `json:"name"`
	Currency        string `json:"currency"`
	AmountMinor     *int64 `json:"amount_minor"`
	BillingInterval string `json:"billing_interval"`
}

type SubscriptionSummary struct {
	Status            string     `json:"status"`
	CurrentPeriodEnd  *time.Time `json:"current_period_end"`
	CancelAtPeriodEnd bool       `json:"cancel_at_period_end"`
	PastDueAt         *time.Time `json:"past_due_at"`
	GraceEndsAt       *time.Time `json:"grace_ends_at"`
	Provider          *string    `json:"provider"`
}

type Status struct {
	DeploymentMode    string               `json:"deployment_mode"`
	BillingConfigured bool                 `json:"billing_configured"`
	Plan              *PlanSummary         `json:"plan"`
	Subscription      *SubscriptionSummary `json:"subscription"`
	Limits            limits.Effective     `json:"limits"`
	Usage             limits.RuntimeUsage  `json:"usage"`
	PortalAvailable   bool                 `json:"portal_available"`
	CheckoutAvailable bool                 `json:"checkout_available"`
}
