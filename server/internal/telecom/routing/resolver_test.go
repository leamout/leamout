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
	trunk      sqlc.Trunk
	endpoints  []sqlc.TrunkEndpoint
	connection sqlc.CarrierConnection
	phone      sqlc.PhoneNumber
	binding    sqlc.GetVoiceBindingByNumberRow
}

func (f *fakeRouteStore) GetTrunk(context.Context, uuid.UUID, uuid.UUID) (sqlc.Trunk, error) {
	return f.trunk, nil
}

func (f *fakeRouteStore) ListOutboundEndpoints(context.Context, uuid.UUID, uuid.UUID) ([]sqlc.TrunkEndpoint, error) {
	return f.endpoints, nil
}

func (f *fakeRouteStore) ResolveInboundCarrier(context.Context, netip.Addr) (sqlc.CarrierConnection, error) {
	return f.connection, nil
}

func (f *fakeRouteStore) GetPhoneNumber(context.Context, uuid.UUID, string) (sqlc.PhoneNumber, error) {
	return f.phone, nil
}

func (f *fakeRouteStore) GetVoiceBinding(context.Context, string) (sqlc.GetVoiceBindingByNumberRow, error) {
	return f.binding, nil
}

func TestResolveOutboundUsesEligibleEndpoint(t *testing.T) {
	organizationID := uuid.New()
	connectionID := uuid.New()
	trunkID := uuid.New()
	endpointID := uuid.New()

	resolver := &Resolver{repo: &fakeRouteStore{
		trunk: sqlc.Trunk{
			ID:                  trunkID,
			OrganizationID:      organizationID,
			CarrierConnectionID: connectionID,
		},
		endpoints: []sqlc.TrunkEndpoint{{
			ID:        endpointID,
			TrunkID:   trunkID,
			Host:      "sip.carrier.example",
			Port:      5060,
			Transport: "udp",
			Priority:  10,
			Weight:    100,
		}},
	}}

	decision, err := resolver.resolveOutbound(context.Background(), OutboundRequest{
		OrganizationID: organizationID,
		TrunkID:        trunkID,
		From:           "+233200000001",
		To:             "+14155550100",
	})
	if err != nil {
		t.Fatalf("resolve outbound: %v", err)
	}
	if decision.CarrierConnectionID != connectionID {
		t.Fatalf("carrier connection = %s, want %s", decision.CarrierConnectionID, connectionID)
	}
	if decision.EndpointID != endpointID {
		t.Fatalf("endpoint = %s, want %s", decision.EndpointID, endpointID)
	}
	if decision.Host != "sip.carrier.example" || decision.Port != 5060 || decision.Transport != "udp" {
		t.Fatalf("unexpected endpoint decision: %+v", decision)
	}
}

func TestResolveOutboundDistributesByWeightAtBestPriority(t *testing.T) {
	organizationID := uuid.New()
	trunkID := uuid.New()
	primaryA := uuid.New()
	primaryB := uuid.New()
	failover := uuid.New()

	resolver := &Resolver{
		repo: &fakeRouteStore{
			trunk: sqlc.Trunk{
				ID:                  trunkID,
				OrganizationID:      organizationID,
				CarrierConnectionID: uuid.New(),
			},
			endpoints: []sqlc.TrunkEndpoint{
				{ID: primaryA, TrunkID: trunkID, Host: "primary-a.example", Port: 5060, Transport: "udp", Priority: 10, Weight: 20},
				{ID: primaryB, TrunkID: trunkID, Host: "primary-b.example", Port: 5060, Transport: "udp", Priority: 10, Weight: 80},
				{ID: failover, TrunkID: trunkID, Host: "failover.example", Port: 5060, Transport: "udp", Priority: 20, Weight: 100},
			},
		},
		pickWeight: func(total int64) (int64, error) {
			if total != 100 {
				t.Fatalf("total weight = %d, want 100", total)
			}
			return 20, nil
		},
	}

	decision, err := resolver.resolveOutbound(context.Background(), OutboundRequest{
		OrganizationID: organizationID,
		TrunkID:        trunkID,
		From:           "+233200000001",
		To:             "+14155550100",
	})
	if err != nil {
		t.Fatalf("resolve outbound: %v", err)
	}
	if decision.EndpointID != primaryB {
		t.Fatalf("endpoint = %s, want weighted primary %s", decision.EndpointID, primaryB)
	}
	if decision.EndpointID == failover {
		t.Fatal("lower-priority failover endpoint received primary traffic")
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
		name string
		pick int64
		want uuid.UUID
	}{
		{name: "first lower boundary", pick: 0, want: firstID},
		{name: "first upper boundary", pick: 1, want: firstID},
		{name: "second lower boundary", pick: 2, want: secondID},
		{name: "second upper boundary", pick: 4, want: secondID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &Resolver{pickWeight: func(int64) (int64, error) { return tt.pick, nil }}
			got, err := resolver.selectOutboundEndpoint(endpoints)
			if err != nil {
				t.Fatalf("select endpoint: %v", err)
			}
			if got.ID != tt.want {
				t.Fatalf("endpoint = %s, want %s", got.ID, tt.want)
			}
		})
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

func TestResolveOutboundReturnsNoRouteWithoutEndpoint(t *testing.T) {
	organizationID := uuid.New()
	trunkID := uuid.New()

	resolver := &Resolver{repo: &fakeRouteStore{
		trunk: sqlc.Trunk{ID: trunkID, OrganizationID: organizationID},
	}}

	_, err := resolver.resolveOutbound(context.Background(), OutboundRequest{
		OrganizationID: organizationID,
		TrunkID:        trunkID,
		From:           "+233200000001",
		To:             "+14155550100",
	})
	if !errors.Is(err, ErrNoRoute) {
		t.Fatalf("error = %v, want %v", err, ErrNoRoute)
	}
}

func TestResolveInbound(t *testing.T) {
	organizationID := uuid.New()
	connectionID := uuid.New()
	phoneNumberID := uuid.New()
	applicationID := uuid.New()

	resolver := &Resolver{repo: &fakeRouteStore{
		connection: sqlc.CarrierConnection{
			ID:             connectionID,
			OrganizationID: organizationID,
		},
		phone: sqlc.PhoneNumber{
			ID:                  phoneNumberID,
			OrganizationID:      organizationID,
			CarrierConnectionID: &connectionID,
			Number:              "+233200000001",
		},
		binding: sqlc.GetVoiceBindingByNumberRow{
			VoiceApplicationID: applicationID,
			PhoneNumberID:      phoneNumberID,
			Number:             "+233200000001",
			OrganizationID:     organizationID,
		},
	}}

	decision, err := resolver.resolveInbound(context.Background(), InboundRequest{
		SourceIP:     "203.0.113.10",
		CalledNumber: "+233200000001",
		CallerNumber: "+14155550100",
	}, netip.MustParseAddr("203.0.113.10"))
	if err != nil {
		t.Fatalf("resolve inbound: %v", err)
	}
	if decision.OrganizationID != organizationID || decision.CarrierConnectionID != connectionID {
		t.Fatalf("unexpected ownership decision: %+v", decision)
	}
	if decision.PhoneNumberID != phoneNumberID || decision.VoiceApplicationID != applicationID {
		t.Fatalf("unexpected application decision: %+v", decision)
	}
}

func TestResolveInboundRejectsCarrierMismatch(t *testing.T) {
	organizationID := uuid.New()
	connectionID := uuid.New()
	otherConnectionID := uuid.New()

	resolver := &Resolver{repo: &fakeRouteStore{
		connection: sqlc.CarrierConnection{
			ID:             connectionID,
			OrganizationID: organizationID,
		},
		phone: sqlc.PhoneNumber{
			ID:                  uuid.New(),
			OrganizationID:      organizationID,
			CarrierConnectionID: &otherConnectionID,
			Number:              "+233200000001",
		},
	}}

	_, err := resolver.resolveInbound(context.Background(), InboundRequest{
		SourceIP:     "203.0.113.10",
		CalledNumber: "+233200000001",
		CallerNumber: "+14155550100",
	}, netip.MustParseAddr("203.0.113.10"))
	if !errors.Is(err, ErrCarrierMismatch) {
		t.Fatalf("error = %v, want %v", err, ErrCarrierMismatch)
	}
}
