package routing

import (
	"context"
	"errors"
	"net/netip"
	"testing"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type fakeRouteStore struct {
	trunk             sqlc.Trunk
	endpoints         []sqlc.TrunkEndpoint
	managedRoutes     []managedRouteCandidate
	connection        sqlc.CarrierConnection
	phone             sqlc.PhoneNumber
	inboundPhone      sqlc.PhoneNumber
	binding           sqlc.GetVoiceBindingByNumberRow
	getTrunkCalls     int
	listManagedCalls  int
}

func (f *fakeRouteStore) GetTrunk(context.Context, uuid.UUID, uuid.UUID) (sqlc.Trunk, error) {
	f.getTrunkCalls++
	return f.trunk, nil
}

func (f *fakeRouteStore) GetCarrierConnection(context.Context, uuid.UUID, uuid.UUID) (sqlc.CarrierConnection, error) {
	connection := f.connection
	if connection.MaxCps == 0 {
		connection.MaxCps = 10
	}
	if connection.MaxConcurrentCalls == 0 {
		connection.MaxConcurrentCalls = 100
	}
	return connection, nil
}

func (f *fakeRouteStore) ListOutboundEndpoints(context.Context, uuid.UUID, uuid.UUID) ([]sqlc.TrunkEndpoint, error) {
	return f.endpoints, nil
}

func (f *fakeRouteStore) ListManagedOutboundRoutes(context.Context) ([]managedRouteCandidate, error) {
	f.listManagedCalls++
	return f.managedRoutes, nil
}

func (f *fakeRouteStore) ResolveInboundCarrier(context.Context, netip.Addr) (sqlc.CarrierConnection, error) {
	return f.connection, nil
}

func (f *fakeRouteStore) GetPhoneNumber(context.Context, uuid.UUID, string) (sqlc.PhoneNumber, error) {
	return f.phone, nil
}

func (f *fakeRouteStore) ResolveInboundPhoneNumber(context.Context, uuid.UUID, string) (sqlc.PhoneNumber, error) {
	if f.inboundPhone.ID != uuid.Nil {
		return f.inboundPhone, nil
	}
	return f.phone, nil
}

func (f *fakeRouteStore) GetVoiceBinding(context.Context, string) (sqlc.GetVoiceBindingByNumberRow, error) {
	return f.binding, nil
}

func uuidPtr(v uuid.UUID) *uuid.UUID { return &v }

func TestResolveOutboundUsesExplicitBYOCTrunk(t *testing.T) {
	organizationID := uuid.New()
	connectionID := uuid.New()
	trunkID := uuid.New()
	endpointID := uuid.New()

	store := &fakeRouteStore{
		trunk: sqlc.Trunk{
			ID:                  trunkID,
			OrganizationID:      uuidPtr(organizationID),
			CarrierConnectionID: connectionID,
		},
		connection: sqlc.CarrierConnection{
			ID:             connectionID,
			OrganizationID: uuidPtr(organizationID),
			Scope:          "organization",
		},
		phone: sqlc.PhoneNumber{
			OrganizationID: organizationID,
			Number: "+233200000001", ProvisioningMode: "byoc", VoiceEnabled: true,
			CarrierConnectionID: &connectionID,
		},
		endpoints: []sqlc.TrunkEndpoint{{
			ID: endpointID, TrunkID: trunkID, Host: "sip.carrier.example", Port: 5060,
			Transport: "udp", Priority: 10, Weight: 100,
		}},
		managedRoutes: []managedRouteCandidate{{EndpointID: uuid.New()}},
	}
	resolver := &Resolver{repo: store, pickWeight: func(int64) (int64, error) { return 0, nil }}

	decision, err := resolver.resolveOutbound(context.Background(), OutboundRequest{
		OrganizationID: organizationID,
		TrunkID:        uuidPtr(trunkID),
		From:           "+233200000001",
		To:             "+14155550100",
	})
	if err != nil {
		t.Fatalf("resolve outbound: %v", err)
	}
	if decision.CarrierConnectionID != connectionID || decision.TrunkID != trunkID || decision.EndpointID != endpointID {
		t.Fatalf("unexpected BYOC decision: %+v", decision)
	}
	if store.listManagedCalls != 0 {
		t.Fatalf("managed route lookup count = %d, want 0 for explicit BYOC", store.listManagedCalls)
	}
}

func TestResolveExplicitBYOCNeverFallsBackToManaged(t *testing.T) {
	organizationID := uuid.New()
	otherOrganizationID := uuid.New()
	trunkID := uuid.New()

	store := &fakeRouteStore{
		trunk: sqlc.Trunk{
			ID:                  trunkID,
			OrganizationID:      uuidPtr(otherOrganizationID),
			CarrierConnectionID: uuid.New(),
		},
		managedRoutes: []managedRouteCandidate{{
			CarrierConnectionID: uuid.New(),
			TrunkID:             uuid.New(),
			EndpointID:          uuid.New(),
			Host:                "managed.carrier.example",
			Port:                5060,
			Transport:           "udp",
			Priority:            10,
			Weight:              100,
		}},
	}
	resolver := &Resolver{repo: store}

	_, err := resolver.resolveOutbound(context.Background(), OutboundRequest{
		OrganizationID: organizationID,
		TrunkID:        uuidPtr(trunkID),
		From:           "+233200000001",
		To:             "+14155550100",
	})
	if !errors.Is(err, ErrNoRoute) {
		t.Fatalf("error = %v, want %v", err, ErrNoRoute)
	}
	if store.listManagedCalls != 0 {
		t.Fatalf("managed route lookup count = %d, want 0 after BYOC failure", store.listManagedCalls)
	}
}

func TestResolveManagedOutboundWithoutTrunk(t *testing.T) {
	organizationID := uuid.New()
	numberConnectionID := uuid.New()
	managedConnectionID := uuid.New()
	managedTrunkID := uuid.New()
	managedEndpointID := uuid.New()

	store := &fakeRouteStore{
		phone: sqlc.PhoneNumber{
			OrganizationID: organizationID, Number: "+233200000001", ProvisioningMode: "managed", VoiceEnabled: true,
			CarrierConnectionID: &numberConnectionID,
		},
		managedRoutes: []managedRouteCandidate{{
			CarrierConnectionID: managedConnectionID,
			MaxCPS:              20,
			MaxConcurrentCalls:  200,
			TrunkID:             managedTrunkID,
			EndpointID:          managedEndpointID,
			Host:                "managed.carrier.example",
			Port:                5060,
			Transport:           "udp",
			Priority:            10,
			Weight:              100,
		}},
	}
	resolver := &Resolver{repo: store, pickWeight: func(int64) (int64, error) { return 0, nil }}

	decision, err := resolver.resolveOutbound(context.Background(), OutboundRequest{
		OrganizationID: organizationID,
		From:           "+233200000001",
		To:             "+14155550100",
	})
	if err != nil {
		t.Fatalf("resolve managed outbound: %v", err)
	}
	if decision.CarrierConnectionID != managedConnectionID || decision.TrunkID != managedTrunkID || decision.EndpointID != managedEndpointID {
		t.Fatalf("unexpected managed decision: %+v", decision)
	}
	if decision.CarrierConnectionID == numberConnectionID {
		t.Fatal("managed route incorrectly required numbering and termination carrier to match")
	}
	if store.getTrunkCalls != 0 {
		t.Fatalf("BYOC trunk lookup count = %d, want 0 for managed routing", store.getTrunkCalls)
	}
}

func TestResolveManagedOutboundReturnsNoRouteWithoutDefault(t *testing.T) {
	organizationID := uuid.New()
	resolver := &Resolver{repo: &fakeRouteStore{
		phone: sqlc.PhoneNumber{
			OrganizationID: organizationID,
			Number:         "+233200000001",
			ProvisioningMode: "managed",
			VoiceEnabled:   true,
		},
	}}
	_, err := resolver.resolveOutbound(context.Background(), OutboundRequest{
		OrganizationID: organizationID, From: "+233200000001", To: "+14155550100",
	})
	if !errors.Is(err, ErrNoRoute) {
		t.Fatalf("error = %v, want %v", err, ErrNoRoute)
	}
}

func TestResolveOutboundRejectsBYOCCallerIdentityFromAnotherCarrier(t *testing.T) {
	organizationID, trunkID := uuid.New(), uuid.New()
	connectionID, otherConnectionID := uuid.New(), uuid.New()
	resolver := &Resolver{repo: &fakeRouteStore{
		trunk:      sqlc.Trunk{ID: trunkID, OrganizationID: uuidPtr(organizationID), CarrierConnectionID: connectionID},
		connection: sqlc.CarrierConnection{ID: connectionID, OrganizationID: uuidPtr(organizationID), Scope: "organization"},
		phone: sqlc.PhoneNumber{
			OrganizationID: organizationID,
			Number: "+233200000001", ProvisioningMode: "byoc", VoiceEnabled: true,
			CarrierConnectionID: &otherConnectionID,
		},
	}}

	_, err := resolver.resolveOutbound(context.Background(), OutboundRequest{
		OrganizationID: organizationID, TrunkID: uuidPtr(trunkID), From: "+233200000001", To: "+14155550100",
	})
	if !errors.Is(err, ErrCallerIdentity) {
		t.Fatalf("error = %v, want %v", err, ErrCallerIdentity)
	}
}

func TestResolveOutboundRejectsManagedNumberOnBYOCTrunk(t *testing.T) {
	organizationID, trunkID := uuid.New(), uuid.New()
	connectionID := uuid.New()
	resolver := &Resolver{repo: &fakeRouteStore{
		trunk:      sqlc.Trunk{ID: trunkID, OrganizationID: uuidPtr(organizationID), CarrierConnectionID: connectionID},
		connection: sqlc.CarrierConnection{ID: connectionID, OrganizationID: uuidPtr(organizationID), Scope: "organization"},
		phone: sqlc.PhoneNumber{
			OrganizationID: organizationID,
			Number: "+233200000001", ProvisioningMode: "managed", VoiceEnabled: true,
			CarrierConnectionID: &connectionID,
		},
	}}

	_, err := resolver.resolveOutbound(context.Background(), OutboundRequest{
		OrganizationID: organizationID, TrunkID: uuidPtr(trunkID), From: "+233200000001", To: "+14155550100",
	})
	if !errors.Is(err, ErrCallerIdentity) {
		t.Fatalf("error = %v, want %v", err, ErrCallerIdentity)
	}
}

func TestSelectOutboundEndpointWeightBoundaries(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	endpoints := []sqlc.TrunkEndpoint{
		{ID: firstID, Priority: 10, Weight: 2},
		{ID: secondID, Priority: 10, Weight: 3},
	}

	tests := []struct {
		pick int64
		want uuid.UUID
	}{
		{pick: 0, want: firstID},
		{pick: 1, want: firstID},
		{pick: 2, want: secondID},
		{pick: 4, want: secondID},
	}
	for _, tt := range tests {
		resolver := &Resolver{pickWeight: func(int64) (int64, error) { return tt.pick, nil }}
		got, err := resolver.selectOutboundEndpoint(endpoints)
		if err != nil {
			t.Fatalf("select endpoint: %v", err)
		}
		if got.ID != tt.want {
			t.Fatalf("endpoint = %s, want %s", got.ID, tt.want)
		}
	}
}

func TestSelectOutboundEndpointFailsOverWhenPrimaryTierIsUnhealthy(t *testing.T) {
	primaryID, failoverID := uuid.New(), uuid.New()
	resolver := &Resolver{pickWeight: func(int64) (int64, error) { return 0, nil }}
	endpoint, err := resolver.selectOutboundEndpoint([]sqlc.TrunkEndpoint{
		{ID: primaryID, Priority: 10, Weight: 100, HealthStatus: "unhealthy"},
		{ID: failoverID, Priority: 20, Weight: 100, HealthStatus: "healthy"},
	})
	if err != nil {
		t.Fatalf("select endpoint: %v", err)
	}
	if endpoint.ID != failoverID {
		t.Fatalf("endpoint = %s, want healthy failover %s", endpoint.ID, failoverID)
	}
}

func TestResolveInboundOrganizationConnection(t *testing.T) {
	organizationID := uuid.New()
	connectionID := uuid.New()
	phoneNumberID := uuid.New()
	applicationID := uuid.New()

	resolver := &Resolver{repo: &fakeRouteStore{
		connection: sqlc.CarrierConnection{ID: connectionID, OrganizationID: uuidPtr(organizationID), Scope: "organization"},
		inboundPhone: sqlc.PhoneNumber{
			ID: phoneNumberID, OrganizationID: organizationID, CarrierConnectionID: &connectionID,
			Number: "+233200000001", ProvisioningMode: "byoc", VoiceEnabled: true,
		},
		binding: sqlc.GetVoiceBindingByNumberRow{
			VoiceApplicationID: applicationID, PhoneNumberID: phoneNumberID,
			Number: "+233200000001", OrganizationID: organizationID,
		},
	}}

	decision, err := resolver.resolveInbound(context.Background(), InboundRequest{
		SourceIP: "203.0.113.10", CalledNumber: "+233200000001", CallerNumber: "+14155550100",
	}, netip.MustParseAddr("203.0.113.10"))
	if err != nil {
		t.Fatalf("resolve inbound: %v", err)
	}
	if decision.OrganizationID != organizationID || decision.CarrierConnectionID != connectionID {
		t.Fatalf("unexpected ownership decision: %+v", decision)
	}
}

func TestResolveInboundPlatformConnectionDerivesTenantFromDID(t *testing.T) {
	organizationID := uuid.New()
	connectionID := uuid.New()
	phoneNumberID := uuid.New()
	applicationID := uuid.New()

	resolver := &Resolver{repo: &fakeRouteStore{
		connection: sqlc.CarrierConnection{ID: connectionID, OrganizationID: nil, Scope: "platform"},
		inboundPhone: sqlc.PhoneNumber{
			ID: phoneNumberID, OrganizationID: organizationID, CarrierConnectionID: &connectionID,
			Number: "+233200000001", ProvisioningMode: "managed", VoiceEnabled: true,
		},
		binding: sqlc.GetVoiceBindingByNumberRow{
			VoiceApplicationID: applicationID, PhoneNumberID: phoneNumberID,
			Number: "+233200000001", OrganizationID: organizationID,
		},
	}}

	decision, err := resolver.resolveInbound(context.Background(), InboundRequest{
		SourceIP: "203.0.113.10", CalledNumber: "+233200000001", CallerNumber: "+14155550100",
	}, netip.MustParseAddr("203.0.113.10"))
	if err != nil {
		t.Fatalf("resolve platform inbound: %v", err)
	}
	if decision.OrganizationID != organizationID {
		t.Fatalf("organization = %s, want DID owner %s", decision.OrganizationID, organizationID)
	}
}

func TestResolveInboundRejectsBYOCNumberOnPlatformConnection(t *testing.T) {
	organizationID := uuid.New()
	connectionID := uuid.New()
	resolver := &Resolver{repo: &fakeRouteStore{
		connection: sqlc.CarrierConnection{ID: connectionID, Scope: "platform"},
		inboundPhone: sqlc.PhoneNumber{
			ID: uuid.New(), OrganizationID: organizationID, CarrierConnectionID: &connectionID,
			Number: "+233200000001", ProvisioningMode: "byoc", VoiceEnabled: true,
		},
	}}

	_, err := resolver.resolveInbound(context.Background(), InboundRequest{
		SourceIP: "203.0.113.10", CalledNumber: "+233200000001", CallerNumber: "+14155550100",
	}, netip.MustParseAddr("203.0.113.10"))
	if !errors.Is(err, ErrTenantMismatch) {
		t.Fatalf("error = %v, want %v", err, ErrTenantMismatch)
	}
}
