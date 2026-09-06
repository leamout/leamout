package edge

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	commercialstate "github.com/leamout/leamout/internal/commercial/state"
)

type fakeStore struct {
	route      Route
	spent      int64
	resolveErr error
	spendErr   error
}

func (f *fakeStore) Resolve(context.Context, Request) (Route, error) {
	return f.route, f.resolveErr
}
func (f *fakeStore) DailyWholesaleSpend(context.Context, uuid.UUID, time.Time) (int64, error) {
	return f.spent, f.spendErr
}

type fakeState struct {
	value commercialstate.OrganizationState
	err   error
}

func (f *fakeState) Resolve(context.Context, uuid.UUID) (commercialstate.OrganizationState, error) {
	return f.value, f.err
}

func TestAdmitAuthorizesManagedIdentityAndRoute(t *testing.T) {
	organizationID, trunkID, connectionID := uuid.New(), uuid.New(), uuid.New()
	store := &fakeStore{route: Route{
		OrganizationID: organizationID, TrunkID: trunkID, CarrierConnectionID: connectionID,
		Host: "wholesale.example", Port: 5061, Transport: "tls",
	}, spent: 99}
	service := NewService(store, &fakeState{value: commercialstate.OrganizationState{
		Standing: commercialstate.StandingActive,
		Features: map[string]bool{ManagedVoiceEntitlement: true},
		Limits:   map[string]int64{ManagedDailySpendLimit: 100},
	}})
	decision, err := service.Admit(context.Background(), Request{To: "+14155550100"})
	if err != nil {
		t.Fatal(err)
	}
	if !decision.Allowed {
		t.Fatal("authorized decision was not marked allowed")
	}
	if decision.TrunkID != trunkID || decision.CarrierConnectionID != connectionID {
		t.Fatalf("decision = %#v", decision)
	}
	if decision.RouteURI != "sip:+14155550100@wholesale.example:5061;transport=tls" {
		t.Fatalf("route URI = %q", decision.RouteURI)
	}
}

func TestAdmitFailsClosedForCommercialState(t *testing.T) {
	route := Route{OrganizationID: uuid.New()}
	tests := []struct {
		name  string
		state commercialstate.OrganizationState
		spent int64
	}{
		{"inactive", commercialstate.OrganizationState{Standing: commercialstate.StandingPastDue}, 0},
		{"feature disabled", commercialstate.OrganizationState{Standing: commercialstate.StandingActive, Limits: map[string]int64{ManagedDailySpendLimit: 10}}, 0},
		{"limit absent", commercialstate.OrganizationState{Standing: commercialstate.StandingActive, Features: map[string]bool{ManagedVoiceEntitlement: true}}, 0},
		{"limit exhausted", commercialstate.OrganizationState{Standing: commercialstate.StandingActive, Features: map[string]bool{ManagedVoiceEntitlement: true}, Limits: map[string]int64{ManagedDailySpendLimit: 10}}, 10},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(&fakeStore{route: route, spent: test.spent}, &fakeState{value: test.state})
			if _, err := service.Admit(context.Background(), Request{}); !errors.Is(err, ErrDenied) {
				t.Fatalf("error = %v, want denied", err)
			}
		})
	}
}

func TestAdmitFailsClosedWhenSpendUnavailable(t *testing.T) {
	service := NewService(&fakeStore{
		route:    Route{OrganizationID: uuid.New()},
		spendErr: errors.New("database unavailable"),
	}, &fakeState{value: commercialstate.OrganizationState{
		Standing: commercialstate.StandingActive,
		Features: map[string]bool{ManagedVoiceEntitlement: true},
		Limits:   map[string]int64{ManagedDailySpendLimit: 100},
	}})
	if _, err := service.Admit(context.Background(), Request{}); err == nil {
		t.Fatal("admission succeeded while dependencies were unavailable")
	}
}

func TestAdmitMapsMissingRouteToDenied(t *testing.T) {
	service := NewService(&fakeStore{resolveErr: pgx.ErrNoRows}, &fakeState{})
	if _, err := service.Admit(context.Background(), Request{}); !errors.Is(err, ErrDenied) {
		t.Fatalf("error = %v, want denied", err)
	}
}

func TestAdmitPreservesRouteLookupFailure(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	service := NewService(&fakeStore{resolveErr: databaseErr}, &fakeState{})
	_, err := service.Admit(context.Background(), Request{})
	if !errors.Is(err, databaseErr) || errors.Is(err, ErrDenied) {
		t.Fatalf("error = %v, want wrapped database error", err)
	}
}
