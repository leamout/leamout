package subscriptions

import "github.com/google/uuid"

const pageSize = 25

type Filter struct {
	Search         string
	Status         string
	Plan           string
	Cancellation   string
	OrganizationID string
	Page           int32
}

type Subscription struct {
	ID                 uuid.UUID
	OrganizationID     uuid.UUID
	OrganizationName   string
	OrganizationStatus string
	PlanID             uuid.UUID
	PlanName           string
	PlanCode           string
	Status             string
	PeriodEnd          string
	CancelAtPeriodEnd  bool
	CreatedAt          string
}

type List struct {
	Subscriptions []Subscription
	Filter        Filter
	HasNext       bool
}

type Detail struct {
	Subscription
	PeriodStart            string
	TrialEndsAt            string
	PastDueAt              string
	GraceEndsAt            string
	CanceledAt             string
	UpdatedAt              string
	Provider               string
	ProviderCustomerID     string
	ProviderSubscriptionID string
	ProviderEventCreatedAt string
}
