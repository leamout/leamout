package edge

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	commercialstate "github.com/leamout/leamout/internal/commercial/state"
)

const (
	ManagedVoiceEntitlement = "voice.managed.enabled"
	ManagedDailySpendLimit  = "voice.managed.daily_spend_micros"
)

var ErrDenied = errors.New("managed SIP call is not authorized")

type Request struct {
	Username string `json:"username"`
	Realm    string `json:"realm"`
	From     string `json:"from"`
	To       string `json:"to"`
}

type Decision struct {
	OrganizationID      uuid.UUID `json:"organization_id"`
	TrunkID             uuid.UUID `json:"trunk_id"`
	CarrierConnectionID uuid.UUID `json:"carrier_connection_id"`
	RouteURI            string    `json:"route_uri"`
}

type Route struct {
	OrganizationID      uuid.UUID
	TrunkID             uuid.UUID
	CarrierConnectionID uuid.UUID
	Host                string
	Port                int32
	Transport           string
}

type store interface {
	Resolve(context.Context, Request) (Route, error)
	DailyWholesaleSpend(context.Context, uuid.UUID, time.Time) (int64, error)
}

type stateResolver interface {
	Resolve(context.Context, uuid.UUID) (commercialstate.OrganizationState, error)
}

type Service struct {
	store store
	state stateResolver
	now   func() time.Time
}

func NewService(store store, state stateResolver) *Service {
	return &Service{store: store, state: state, now: time.Now}
}

func (s *Service) Admit(ctx context.Context, req Request) (Decision, error) {
	if s.store == nil || s.state == nil {
		return Decision{}, fmt.Errorf("managed SIP admission dependencies are unavailable")
	}
	route, err := s.store.Resolve(ctx, req)
	if err != nil {
		return Decision{}, fmt.Errorf("resolve managed SIP route: %w", err)
	}
	state, err := s.state.Resolve(ctx, route.OrganizationID)
	if err != nil {
		return Decision{}, fmt.Errorf("resolve managed SIP commercial state: %w", err)
	}
	limit, ok := state.Limit(ManagedDailySpendLimit)
	if state.Standing != commercialstate.StandingActive || !state.Enabled(ManagedVoiceEntitlement) || !ok || limit <= 0 {
		return Decision{}, ErrDenied
	}
	now := s.now().UTC()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	spent, err := s.store.DailyWholesaleSpend(ctx, route.OrganizationID, day)
	if err != nil {
		return Decision{}, fmt.Errorf("resolve managed SIP wholesale spend: %w", err)
	}
	if spent >= limit {
		return Decision{}, ErrDenied
	}
	return Decision{
		OrganizationID: route.OrganizationID, TrunkID: route.TrunkID,
		CarrierConnectionID: route.CarrierConnectionID,
		RouteURI:            fmt.Sprintf("sip:%s@%s:%d;transport=%s", req.To, route.Host, route.Port, route.Transport),
	}, nil
}
