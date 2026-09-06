package trunks

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	commercialstate "github.com/leamout/leamout/internal/commercial/state"
	"github.com/leamout/leamout/pkg/hasher"
)

type fakeManagedSIPState struct {
	state commercialstate.OrganizationState
	err   error
}

func (f fakeManagedSIPState) Resolve(context.Context, uuid.UUID) (commercialstate.OrganizationState, error) {
	return f.state, f.err
}

func TestCreateManagedTrunkFailsClosedWithoutAuthority(t *testing.T) {
	service := NewService(nil)

	_, err := service.Create(context.Background(), uuid.New(), CreateRequest{
		Type: ProvisioningModeManaged,
		Name: "Leamout managed",
	})
	if err == nil {
		t.Fatal("managed trunk creation succeeded without hosted managed SIP authority")
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

func TestManagedSIPAuthorityRequiresCommercialEntitlement(t *testing.T) {
	organizationID := uuid.New()
	service := NewService(nil)
	if err := service.SetManagedSIP(ManagedSIPConfig{
		Enabled: true, Host: "sip.leamout.com", Port: 5061, Transport: "tls", Realm: "sip.leamout.com",
	}, fakeManagedSIPState{state: commercialstate.OrganizationState{
		OrganizationID: organizationID,
		Standing:       commercialstate.StandingActive,
		Features:       map[string]bool{},
	}}); err != nil {
		t.Fatalf("configure managed SIP: %v", err)
	}
	if err := service.authorizeManagedSIP(context.Background(), organizationID); err == nil {
		t.Fatal("managed SIP authority accepted an organization without voice.managed.enabled")
	}
}

func TestManagedSIPCredentialIsOneWayDigestMaterial(t *testing.T) {
	organizationID := uuid.New()
	service := NewService(nil)
	if err := service.SetManagedSIP(ManagedSIPConfig{
		Enabled: true, Host: "SIP.LEAMOUT.COM", Port: 5061, Transport: "TLS", Realm: "sip.leamout.com",
	}, fakeManagedSIPState{state: commercialstate.OrganizationState{
		OrganizationID: organizationID,
		Standing:       commercialstate.StandingActive,
		Features:       map[string]bool{ManagedVoiceEntitlement: true},
	}}); err != nil {
		t.Fatalf("configure managed SIP: %v", err)
	}
	if err := service.authorizeManagedSIP(context.Background(), organizationID); err != nil {
		t.Fatalf("authorize managed SIP: %v", err)
	}

	credential, ha1, err := service.newManagedSIPCredential()
	if err != nil {
		t.Fatalf("generate managed SIP credential: %v", err)
	}
	if credential.Host != "sip.leamout.com" || credential.Transport != "tls" || credential.Port != 5061 {
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
