package entitlements

import "time"

// Feature identifies a commercial capability that can be enabled for a customer.
type Feature string

// Kind identifies how an entitlement value is interpreted.
type Kind string

const (
	KindFeature Kind = "feature"
	KindLimit   Kind = "limit"
)

// Entitlement is a feature switch or capacity limit granted by a plan or an
// explicit organization/license override. Exactly one value is populated.
type Entitlement struct {
	Key       string
	Kind      Kind
	Enabled   *bool
	Limit     *int64
	StartsAt  *time.Time
	ExpiresAt *time.Time
}

// ActiveAt reports whether the entitlement is valid at the supplied instant.
func (e Entitlement) ActiveAt(at time.Time) bool {
	return (e.StartsAt == nil || !at.Before(*e.StartsAt)) &&
		(e.ExpiresAt == nil || at.Before(*e.ExpiresAt))
}

// EntitlementSet is the effective set of commercial capabilities and limits.
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
