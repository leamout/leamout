package state

import (
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

// Standing describes the effective commercial access posture of an organization.
type Standing string

const (
	StandingUnsubscribed Standing = "unsubscribed"
	StandingActive       Standing = "active"
	StandingPastDue      Standing = "past_due"
)

var ErrOrganizationIDRequired = apperror.NewBadRequest("organization id is required")

// OrganizationState is the resolved commercial state consumed by operational modules.
type OrganizationState struct {
	OrganizationID uuid.UUID
	Standing       Standing
	SubscriptionID *uuid.UUID
	PlanID         *uuid.UUID
	Features       map[string]bool
	Limits         map[string]int64
	EffectiveAt    time.Time
	NextChangeAt   *time.Time
}

type organizationStateResponse struct {
	OrganizationID uuid.UUID        `json:"organization_id"`
	Standing       Standing         `json:"standing"`
	SubscriptionID *uuid.UUID       `json:"subscription_id,omitempty"`
	PlanID         *uuid.UUID       `json:"plan_id,omitempty"`
	Features       map[string]bool  `json:"features"`
	Limits         map[string]int64 `json:"limits"`
	EffectiveAt    time.Time        `json:"effective_at"`
	NextChangeAt   *time.Time       `json:"next_change_at,omitempty"`
}

func newOrganizationStateResponse(state OrganizationState) organizationStateResponse {
	return organizationStateResponse(state)
}

func (s OrganizationState) Enabled(feature string) bool {
	return s.Features[feature]
}

func (s OrganizationState) Limit(name string) (int64, bool) {
	value, ok := s.Limits[name]
	return value, ok
}
