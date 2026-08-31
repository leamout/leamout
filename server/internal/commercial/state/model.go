package state

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Standing describes the effective commercial access posture of an organization.
type Standing string

const (
	StandingUnsubscribed Standing = "unsubscribed"
	StandingActive       Standing = "active"
	StandingPastDue      Standing = "past_due"
)

var (
	ErrOrganizationIDRequired = errors.New("organization id is required")
	ErrUnavailable            = errors.New("commercial state unavailable")
)

// OrganizationState is the resolved commercial state consumed by operational modules.
type OrganizationState struct {
	OrganizationID uuid.UUID
	Standing       Standing
	SubscriptionID *uuid.UUID
	PlanID         *uuid.UUID
	Features       map[string]bool
	Limits         map[string]int64
	EffectiveAt    time.Time
}

func (s OrganizationState) Enabled(feature string) bool {
	return s.Features[feature]
}

func (s OrganizationState) Limit(name string) (int64, bool) {
	value, ok := s.Limits[name]
	return value, ok
}
