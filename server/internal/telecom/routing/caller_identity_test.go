package routing

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

func TestAuthorizeBYOCCallerIdentityRequiresSameCarrier(t *testing.T) {
	organizationID := uuid.New()
	carrierConnectionID := uuid.New()
	otherCarrierConnectionID := uuid.New()

	caller := sqlc.PhoneNumber{
		OrganizationID:      organizationID,
		ProvisioningMode:    provisioningModeBYOC,
		CarrierConnectionID: &otherCarrierConnectionID,
		VoiceEnabled:        true,
		Status:              phoneNumberStatusActive,
	}

	if err := authorizeBYOCCallerIdentity(caller, organizationID, carrierConnectionID); !errors.Is(err, ErrCallerIdentity) {
		t.Fatalf("error = %v, want %v", err, ErrCallerIdentity)
	}

	caller.CarrierConnectionID = &carrierConnectionID
	if err := authorizeBYOCCallerIdentity(caller, organizationID, carrierConnectionID); err != nil {
		t.Fatalf("authorize BYOC caller: %v", err)
	}
}

func TestAuthorizeManagedCallerIdentityAllowsDifferentTerminationProvider(t *testing.T) {
	organizationID := uuid.New()
	didwwProviderID := uuid.New()
	commpeakProviderID := uuid.New()

	caller := sqlc.PhoneNumber{
		OrganizationID:   organizationID,
		ProvisioningMode: provisioningModeManaged,
		ProviderID:       &didwwProviderID,
		VoiceEnabled:     true,
		Status:           phoneNumberStatusActive,
	}

	// The termination provider is intentionally different from the numbering
	// provider. Managed caller authorization does not receive or compare it.
	if didwwProviderID == commpeakProviderID {
		t.Fatal("test requires distinct numbering and termination providers")
	}
	if err := authorizeManagedCallerIdentity(caller, organizationID); err != nil {
		t.Fatalf("authorize DIDWW caller for CommPeak termination: %v", err)
	}
}

func TestAuthorizeManagedCallerIdentityRejectsBYOCNumber(t *testing.T) {
	organizationID := uuid.New()
	caller := sqlc.PhoneNumber{
		OrganizationID:   organizationID,
		ProvisioningMode: provisioningModeBYOC,
		VoiceEnabled:     true,
		Status:           phoneNumberStatusActive,
	}

	if err := authorizeManagedCallerIdentity(caller, organizationID); !errors.Is(err, ErrCallerIdentity) {
		t.Fatalf("error = %v, want %v", err, ErrCallerIdentity)
	}
}

func TestAuthorizeManagedCallerIdentityRejectsDisabledNumber(t *testing.T) {
	organizationID := uuid.New()
	caller := sqlc.PhoneNumber{
		OrganizationID:   organizationID,
		ProvisioningMode: provisioningModeManaged,
		VoiceEnabled:     true,
		Status:           "disabled",
	}

	if err := authorizeManagedCallerIdentity(caller, organizationID); !errors.Is(err, ErrCallerIdentity) {
		t.Fatalf("error = %v, want %v", err, ErrCallerIdentity)
	}
}

func TestAuthorizeManagedCallerIdentityRejectsOtherOrganization(t *testing.T) {
	caller := sqlc.PhoneNumber{
		OrganizationID:   uuid.New(),
		ProvisioningMode: provisioningModeManaged,
		VoiceEnabled:     true,
		Status:           phoneNumberStatusActive,
	}

	if err := authorizeManagedCallerIdentity(caller, uuid.New()); !errors.Is(err, ErrCallerIdentity) {
		t.Fatalf("error = %v, want %v", err, ErrCallerIdentity)
	}
}
