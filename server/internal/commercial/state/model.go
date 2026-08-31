package state

import "time"

// OrganizationState is the resolved commercial state consumed by operational modules.
type OrganizationState struct {
	OrganizationID string
	SubscriptionID *string
	PlanID         *string
	LicenseID      *string
	Features       map[string]bool
	Limits         map[string]int64
	EffectiveAt    time.Time
	ExpiresAt      *time.Time
}
