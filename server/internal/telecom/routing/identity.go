package routing

import (
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

const phoneNumberStatusActive = "active"

func callerIdentityStatusAllowed(status string) bool {
	return status == "" || status == phoneNumberStatusActive
}

// authorizeBYOCCallerIdentity preserves carrier affinity for customer-selected
// connectivity. The number must belong to the organization, be active and
// voice-enabled, and be assigned to the exact organization carrier connection
// selected by the requested trunk.
func authorizeBYOCCallerIdentity(
	caller sqlc.PhoneNumber,
	organizationID uuid.UUID,
	carrierConnectionID uuid.UUID,
) error {
	if caller.OrganizationID != organizationID ||
		!callerIdentityStatusAllowed(caller.Status) ||
		!caller.VoiceEnabled ||
		caller.CarrierConnectionID == nil ||
		*caller.CarrierConnectionID != carrierConnectionID {
		return ErrCallerIdentity
	}
	return nil
}

// authorizeManagedCallerIdentity validates ownership and number state only.
// Cloud Managed termination chooses its platform carrier independently from
// how the caller identity entered Leamout.
func authorizeManagedCallerIdentity(
	caller sqlc.PhoneNumber,
	organizationID uuid.UUID,
) error {
	if caller.OrganizationID != organizationID ||
		!callerIdentityStatusAllowed(caller.Status) ||
		!caller.VoiceEnabled {
		return ErrCallerIdentity
	}
	return nil
}
