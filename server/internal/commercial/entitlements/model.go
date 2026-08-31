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
	ID             uuid.UUID
	PlanID         *uuid.UUID
	OrganizationID *uuid.UUID
	LicenseID      *uuid.UUID
	Key            string
	Kind           Kind
	Enabled        *bool
	LimitValue     *int64
	StartsAt       *time.Time
	ExpiresAt      *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CreateInput describes the value and optional activation window for an entitlement.
// The owning scope is supplied by the service method rather than embedded here.
type CreateInput struct {
	Key        string
	Kind       Kind
	Enabled    *bool
	LimitValue *int64
	StartsAt   *time.Time
	ExpiresAt  *time.Time
}

// EntitlementSet is the resolved set of commercial capabilities and limits.
type EntitlementSet struct {
	Features map[Feature]bool
	Limits   map[string]int64
}

func (e EntitlementSet) Enabled(feature Feature) bool {
	return e.Features[feature]
}

func (e EntitlementSet) Limit(name string) (int64, bool) {
	value, ok := e.Limits[name]
	return value, ok
}
