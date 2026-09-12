package edge

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrDenied = errors.New("managed SIP call is not authorized")

type Request struct {
	Username string `json:"username"`
	Realm    string `json:"realm"`
	From     string `json:"from"`
	To       string `json:"to"`
}

type Decision struct {
	Allowed             bool      `json:"allowed"`
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
}

type Service struct{ store store }

func NewService(store store) *Service { return &Service{store: store} }

// Admit validates the tenant-owned managed SIP route. Prepaid authorization is
// enforced at the provider-obligation boundary by the managed call workflow.
func (s *Service) Admit(ctx context.Context, req Request) (Decision, error) {
	if s.store == nil {
		return Decision{}, fmt.Errorf("managed SIP admission dependencies are unavailable")
	}
	route, err := s.store.Resolve(ctx, req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Decision{}, ErrDenied
		}
		return Decision{}, fmt.Errorf("resolve managed SIP route: %w", err)
	}
	return Decision{
		Allowed:             true,
		OrganizationID:      route.OrganizationID,
		TrunkID:             route.TrunkID,
		CarrierConnectionID: route.CarrierConnectionID,
		RouteURI:            fmt.Sprintf("sip:%s@%s:%d;transport=%s", req.To, route.Host, route.Port, route.Transport),
	}, nil
}
