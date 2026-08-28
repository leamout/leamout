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

func TestResolveOutboundUsesFirstEligibleEndpoint(t *testing.T) {
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
