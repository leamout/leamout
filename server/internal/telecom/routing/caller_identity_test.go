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

func TestAuthorizeManagedCallerIdentityAllowsProviderOwnedNumber(t *testing.T) {
	organizationID := uuid.New()
	providerID := uuid.New()
	caller := sqlc.PhoneNumber{
		OrganizationID: organizationID,
		ProviderID:     &providerID,
		VoiceEnabled:   true,
		Status:         phoneNumberStatusActive,
	}

	if err := authorizeManagedCallerIdentity(caller, organizationID); err != nil {
		t.Fatalf("authorize provider-owned caller: %v", err)
	}
}

func TestAuthorizeManagedCallerIdentityAllowsCustomerNumber(t *testing.T) {
	organizationID := uuid.New()
	caller := sqlc.PhoneNumber{
		OrganizationID: organizationID,
		VoiceEnabled:   true,
		Status:         phoneNumberStatusActive,
	}

	if err := authorizeManagedCallerIdentity(caller, organizationID); err != nil {
		t.Fatalf("authorize customer caller: %v", err)
	}
}

func TestAuthorizeManagedCallerIdentityRejectsDisabledNumber(t *testing.T) {
	organizationID := uuid.New()
	caller := sqlc.PhoneNumber{
		OrganizationID: organizationID,
		VoiceEnabled:   true,
		Status:         "disabled",
	}

	if err := authorizeManagedCallerIdentity(caller, organizationID); !errors.Is(err, ErrCallerIdentity) {
		t.Fatalf("error = %v, want %v", err, ErrCallerIdentity)
	}
}

func TestAuthorizeManagedCallerIdentityRejectsOtherOrganization(t *testing.T) {
	caller := sqlc.PhoneNumber{
		OrganizationID: uuid.New(),
		VoiceEnabled:   true,
		Status:         phoneNumberStatusActive,
	}

	if err := authorizeManagedCallerIdentity(caller, uuid.New()); !errors.Is(err, ErrCallerIdentity) {
		t.Fatalf("error = %v, want %v", err, ErrCallerIdentity)
	}
}
