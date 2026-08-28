package routing

import (
	"context"
	"errors"
	"net/netip"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type routeStore interface {
	GetTrunk(context.Context, uuid.UUID, uuid.UUID) (sqlc.Trunk, error)
	ListOutboundEndpoints(context.Context, uuid.UUID, uuid.UUID) ([]sqlc.TrunkEndpoint, error)
	ResolveInboundCarrier(context.Context, netip.Addr) (sqlc.CarrierConnection, error)
	GetPhoneNumber(context.Context, uuid.UUID, string) (sqlc.PhoneNumber, error)
	GetVoiceBinding(context.Context, string) (sqlc.GetVoiceBindingByNumberRow, error)
}

type Resolver struct {
	repo routeStore
}

func NewResolver(repo *Repository) *Resolver {
	return &Resolver{repo: repo}
}

func (r *Resolver) resolveOutbound(
	ctx context.Context,
	req OutboundRequest,
) (OutboundDecision, error) {
	trunk, err := r.repo.GetTrunk(ctx, req.OrganizationID, req.TrunkID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OutboundDecision{}, ErrNoRoute
		}
		return OutboundDecision{}, err
	}

	endpoints, err := r.repo.ListOutboundEndpoints(ctx, req.OrganizationID, req.TrunkID)
	if err != nil {
		return OutboundDecision{}, err
	}
	if len(endpoints) == 0 {
		return OutboundDecision{}, ErrNoRoute
	}

	endpoint := endpoints[0]
	return OutboundDecision{
		OrganizationID:      req.OrganizationID,
		CarrierConnectionID: trunk.CarrierConnectionID,
		TrunkID:             trunk.ID,
		EndpointID:          endpoint.ID,
		Host:                endpoint.Host,
		Port:                endpoint.Port,
		Transport:           endpoint.Transport,
		From:                req.From,
		To:                  req.To,
	}, nil
}

func (r *Resolver) resolveInbound(
	ctx context.Context,
	req InboundRequest,
	sourceIP netip.Addr,
) (InboundDecision, error) {
	connection, err := r.repo.ResolveInboundCarrier(ctx, sourceIP)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InboundDecision{}, ErrNoRoute
		}
		return InboundDecision{}, err
	}

	phoneNumber, err := r.repo.GetPhoneNumber(ctx, connection.OrganizationID, req.CalledNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InboundDecision{}, ErrNoRoute
		}
		return InboundDecision{}, err
	}

	if phoneNumber.CarrierConnectionID == nil || *phoneNumber.CarrierConnectionID != connection.ID {
		return InboundDecision{}, ErrCarrierMismatch
	}

	binding, err := r.repo.GetVoiceBinding(ctx, req.CalledNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InboundDecision{}, ErrNoRoute
		}
		return InboundDecision{}, err
	}
	if binding.OrganizationID != connection.OrganizationID {
		return InboundDecision{}, ErrTenantMismatch
	}

	return InboundDecision{
		OrganizationID:      connection.OrganizationID,
		CarrierConnectionID: connection.ID,
		PhoneNumberID:       phoneNumber.ID,
		VoiceApplicationID:  binding.VoiceApplicationID,
		CalledNumber:        req.CalledNumber,
		CallerNumber:        req.CallerNumber,
	}, nil
}
