package edge

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type fakeStore struct {
	route      Route
	resolveErr error
}

func (f *fakeStore) Resolve(context.Context, Request) (Route, error) {
	return f.route, f.resolveErr
}

func TestAdmitAuthorizesTenantOwnedManagedRoute(t *testing.T) {
	organizationID, trunkID, connectionID := uuid.New(), uuid.New(), uuid.New()
	store := &fakeStore{route: Route{
		OrganizationID: organizationID, TrunkID: trunkID, CarrierConnectionID: connectionID,
		Host: "wholesale.example", Port: 5061, Transport: "tls",
	}}
	service := NewService(store)
	decision, err := service.Admit(context.Background(), Request{To: "+14155550100"})
	if err != nil {
		t.Fatal(err)
	}
	if !decision.Allowed {
		t.Fatal("authorized decision was not marked allowed")
	}
	if decision.OrganizationID != organizationID || decision.TrunkID != trunkID || decision.CarrierConnectionID != connectionID {
		t.Fatalf("decision = %#v", decision)
	}
	if decision.RouteURI != "sip:+14155550100@wholesale.example:5061;transport=tls" {
		t.Fatalf("route URI = %q", decision.RouteURI)
	}
}

func TestAdmitFailsClosedWithoutStore(t *testing.T) {
	service := NewService(nil)
	if _, err := service.Admit(context.Background(), Request{}); err == nil {
		t.Fatal("admission succeeded without route store")
	}
}

func TestAdmitMapsMissingRouteToDenied(t *testing.T) {
	service := NewService(&fakeStore{resolveErr: pgx.ErrNoRows})
	if _, err := service.Admit(context.Background(), Request{}); !errors.Is(err, ErrDenied) {
		t.Fatalf("error = %v, want denied", err)
	}
}

func TestAdmitPreservesRouteLookupFailure(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	service := NewService(&fakeStore{resolveErr: databaseErr})
	_, err := service.Admit(context.Background(), Request{})
	if !errors.Is(err, databaseErr) || errors.Is(err, ErrDenied) {
		t.Fatalf("error = %v, want wrapped database error", err)
	}
}
