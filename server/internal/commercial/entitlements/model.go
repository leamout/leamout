package entitlements

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Kind identifies the value shape carried by an entitlement.
type Kind string

const (
	KindFeature Kind = "feature"
	KindLimit   Kind = "limit"
)

// Feature identifies a commercial capability that can be enabled or disabled.
type Feature string

var (
	ErrEntitlementConflict     = errors.New("entitlement key already exists in scope")
	ErrInvalidEntitlement      = errors.New("invalid entitlement")
	ErrKeyRequired             = errors.New("entitlement key is required")
	ErrInvalidKey              = errors.New("entitlement key must not contain whitespace")
	ErrInvalidKind             = errors.New("invalid entitlement kind")
	ErrFeatureValueRequired    = errors.New("feature entitlement requires enabled")
	ErrLimitValueRequired      = errors.New("limit entitlement requires limit_value")
	ErrInvalidLimit            = errors.New("entitlement limit must be non-negative")
	ErrInvalidPeriod           = errors.New("invalid entitlement period")
	ErrScopeUnavailable        = errors.New("entitlement scope is unavailable")
	ErrSubscriptionUnavailable = errors.New("organization has no current subscription")
	ErrKindMismatch            = errors.New("entitlement override kind does not match inherited kind")
	ErrPlanIDRequired          = errors.New("plan id is required")
	ErrOrganizationIDRequired  = errors.New("organization id is required")
	ErrLicenseIDRequired       = errors.New("license id is required")
	ErrEntitlementIDRequired   = errors.New("entitlement id is required")
)

// Entitlement is one durable feature or limit attached to exactly one scope.
type Entitlement struct {
	ID             uuid.UUID  `json:"id"`
	PlanID         *uuid.UUID `json:"plan_id,omitempty"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty"`
	LicenseID      *uuid.UUID `json:"license_id,omitempty"`
	Key            string     `json:"key"`
	Kind           Kind       `json:"kind"`
	Enabled        *bool      `json:"enabled,omitempty"`
	LimitValue     *int64     `json:"limit_value,omitempty"`
	StartsAt       *time.Time `json:"starts_at,omitempty"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CreateInput describes the value and optional activation window for an entitlement.
// The owning scope is supplied by the service method rather than embedded here.
type CreateInput struct {
	Key        string     `json:"key"`
	Kind       Kind       `json:"kind"`
	Enabled    *bool      `json:"enabled,omitempty"`
	LimitValue *int64     `json:"limit_value,omitempty"`
	StartsAt   *time.Time `json:"starts_at,omitempty"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
}

// EntitlementSet is the resolved set of commercial capabilities and limits.
type EntitlementSet struct {
	Features map[Feature]bool
	Limits   map[string]int64
}

// Resolution carries effective entitlements plus the earliest known entitlement
// boundary that can require the commercial state to be resolved again.
type Resolution struct {
	Set          EntitlementSet
	NextChangeAt *time.Time
}

func (e EntitlementSet) Enabled(feature Feature) bool {
	return e.Features[feature]
}

func (e EntitlementSet) Limit(name string) (int64, bool) {
	value, ok := e.Limits[name]
	return value, ok
}
