package routing

import (
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

const (
	provisioningModeBYOC    = "byoc"
	provisioningModeManaged = "managed"
	phoneNumberStatusActive = "active"
)

// authorizeBYOCCallerIdentity preserves carrier affinity for customer-owned
// connectivity. A BYOC caller ID is valid only when the number belongs to the
// calling organization, is active and voice-enabled, is itself BYOC, and is
// assigned to the exact carrier connection selected by the requested trunk.
func authorizeBYOCCallerIdentity(
	caller sqlc.PhoneNumber,
	organizationID uuid.UUID,
	carrierConnectionID uuid.UUID,
) error {
	if caller.OrganizationID != organizationID ||
		caller.ProvisioningMode != provisioningModeBYOC ||
		caller.Status != phoneNumberStatusActive ||
		!caller.VoiceEnabled ||
		caller.CarrierConnectionID == nil ||
		*caller.CarrierConnectionID != carrierConnectionID {
		return ErrCallerIdentity
	}
	return nil
}

// authorizeManagedCallerIdentity deliberately has no termination-carrier
// argument. Managed caller identity is authorized by organization ownership
// and managed-number state, not by affinity between the numbering provider and
// the platform carrier selected for termination. This permits, for example, a
// DIDWW-managed DID to be presented on a call terminated through CommPeak.
func authorizeManagedCallerIdentity(
	caller sqlc.PhoneNumber,
	organizationID uuid.UUID,
) error {
	if caller.OrganizationID != organizationID ||
		caller.ProvisioningMode != provisioningModeManaged ||
		caller.Status != phoneNumberStatusActive ||
		!caller.VoiceEnabled {
		return ErrCallerIdentity
	}
	return nil
}
