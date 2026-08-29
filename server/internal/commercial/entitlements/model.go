package entitlements

// Feature identifies a commercial capability that can be enabled for a customer.
type Feature string

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
