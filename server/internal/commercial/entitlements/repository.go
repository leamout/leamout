package entitlements

import "context"

// Repository resolves durable entitlement records within an organization boundary.
type Repository interface {
	EffectiveForOrganization(context.Context, string) (EntitlementSet, error)
	ForLicense(context.Context, string, string) (EntitlementSet, error)
}
