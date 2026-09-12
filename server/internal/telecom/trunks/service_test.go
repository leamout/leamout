package trunks

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/hasher"
)

func TestCreateManagedTrunkRequiresDatabase(t *testing.T) {
	service := NewService(nil)

	_, err := service.Create(context.Background(), uuid.New(), CreateRequest{
		Type: ProvisioningModeManaged,
		Name: "Leamout managed",
	})
	if err == nil {
		t.Fatal("managed trunk creation succeeded without database")
	}
}

func TestCreateManagedTrunkRejectsCarrierConnection(t *testing.T) {
	service := NewService(nil)
	connectionID := uuid.New()

	_, err := service.Create(context.Background(), uuid.New(), CreateRequest{
		Type:                ProvisioningModeManaged,
		CarrierConnectionID: &connectionID,
		Name:                "Leamout managed",
	})
	if err == nil {
		t.Fatal("managed trunk accepted a customer carrier_connection_id")
	}
}

func TestManagedSIPCredentialIsOneWayDigestMaterial(t *testing.T) {
	service := NewService(nil)

	credential, ha1, err := service.newManagedSIPCredential()
	if err != nil {
		t.Fatalf("generate managed SIP credential: %v", err)
	}
	if credential.Host != ManagedSIPHost || credential.Transport != ManagedSIPTransport || credential.Port != ManagedSIPPort || credential.Realm != ManagedSIPRealm {
		t.Fatalf("unexpected managed SIP endpoint: %+v", credential)
	}
	if !strings.HasPrefix(credential.Username, "lm_sip_") || !strings.HasPrefix(credential.Password, "lm_sip_") {
		t.Fatalf("unexpected credential shape: username=%q password=%q", credential.Username, credential.Password)
	}
	if ha1 == credential.Password || ha1 == "" {
		t.Fatal("stored HA1 must be non-empty and must not equal the plaintext password")
	}
	if want := hasher.ComputeHA1MD5(credential.Username, credential.Realm, credential.Password); ha1 != want {
		t.Fatalf("HA1 = %q, want %q", ha1, want)
	}

	second, secondHA1, err := service.newManagedSIPCredential()
	if err != nil {
		t.Fatalf("generate second managed SIP credential: %v", err)
	}
	if credential.Username == second.Username || credential.Password == second.Password || ha1 == secondHA1 {
		t.Fatal("independent managed SIP credentials must not reuse identity or secret material")
	}
}
