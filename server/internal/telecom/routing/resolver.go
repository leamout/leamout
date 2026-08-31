package routing

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/netip"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type routeStore interface {
	GetTrunk(context.Context, uuid.UUID, uuid.UUID) (sqlc.Trunk, error)
	GetCarrierConnection(context.Context, uuid.UUID, uuid.UUID) (sqlc.CarrierConnection, error)
	ListOutboundEndpoints(context.Context, uuid.UUID, uuid.UUID) ([]sqlc.TrunkEndpoint, error)
	ResolveInboundCarrier(context.Context, netip.Addr) (sqlc.CarrierConnection, error)
	GetPhoneNumber(context.Context, uuid.UUID, string) (sqlc.PhoneNumber, error)
	GetVoiceBinding(context.Context, string) (sqlc.GetVoiceBindingByNumberRow, error)
}

type Resolver struct {
	repo       routeStore
	pickWeight func(int64) (int64, error)
	metrics    interface {
		EndpointSelection(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, bool)
	}
}

func (r *Resolver) SetMetrics(metrics interface {
	EndpointSelection(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, bool)
}) {
	r.metrics = metrics
}

func NewResolver(repo *Repository) *Resolver {
	return &Resolver{repo: repo, pickWeight: secureWeightPick}
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
	connection, err := r.repo.GetCarrierConnection(ctx, req.OrganizationID, trunk.CarrierConnectionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OutboundDecision{}, ErrNoRoute
		}
		return OutboundDecision{}, err
	}
	caller, err := r.repo.GetPhoneNumber(ctx, req.OrganizationID, req.From)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OutboundDecision{}, ErrCallerIdentity
		}
		return OutboundDecision{}, err
	}
	if !caller.VoiceEnabled || caller.CarrierConnectionID == nil || *caller.CarrierConnectionID != trunk.CarrierConnectionID {
		return OutboundDecision{}, ErrCallerIdentity
	}

	endpoints, err := r.repo.ListOutboundEndpoints(ctx, req.OrganizationID, req.TrunkID)
	if err != nil {
		return OutboundDecision{}, err
	}
	if len(endpoints) == 0 {
		return OutboundDecision{}, ErrNoRoute
	}

	endpoint, err := r.selectOutboundEndpoint(endpoints)
	if err != nil {
		return OutboundDecision{}, err
	}
	if r.metrics != nil {
		failover := len(endpoints) > 0 && endpoint.Priority > endpoints[0].Priority
		r.metrics.EndpointSelection(ctx, trunk.CarrierConnectionID, trunk.ID, endpoint.ID, failover)
	}
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
		MaxCPS:              connection.MaxCps,
		MaxConcurrentCalls:  connection.MaxConcurrentCalls,
		MaxDailyMinutes:     connection.MaxDailyMinutes,
	}, nil
}

// selectOutboundEndpoint restricts routing to the best available priority and
// distributes calls by weight within that priority. Lower-priority endpoints
// are failover targets and must not receive traffic while a better priority is
// eligible.
func (r *Resolver) selectOutboundEndpoint(endpoints []sqlc.TrunkEndpoint) (sqlc.TrunkEndpoint, error) {
	if len(endpoints) == 0 {
		return sqlc.TrunkEndpoint{}, ErrNoRoute
	}

	var priority int32
	prioritySet := false
	var total int64
	eligible := make([]sqlc.TrunkEndpoint, 0, len(endpoints))
	for _, endpoint := range endpoints {
		if endpoint.HealthStatus == "unhealthy" {
			continue
		}
		if !prioritySet {
			priority = endpoint.Priority
			prioritySet = true
		}
		if endpoint.Priority != priority {
			break
		}
		if endpoint.Weight <= 0 {
			return sqlc.TrunkEndpoint{}, fmt.Errorf("endpoint %s has invalid weight %d", endpoint.ID, endpoint.Weight)
		}
		total += int64(endpoint.Weight)
		eligible = append(eligible, endpoint)
	}
	if len(eligible) == 0 {
		return sqlc.TrunkEndpoint{}, ErrNoRoute
	}

	picker := r.pickWeight
	if picker == nil {
		picker = secureWeightPick
	}
	pick, err := picker(total)
	if err != nil {
		return sqlc.TrunkEndpoint{}, fmt.Errorf("select weighted carrier endpoint: %w", err)
	}
	if pick < 0 || pick >= total {
		return sqlc.TrunkEndpoint{}, fmt.Errorf("weighted carrier endpoint selection returned %d outside [0,%d)", pick, total)
	}

	for _, endpoint := range eligible {
		if pick < int64(endpoint.Weight) {
			return endpoint, nil
		}
		pick -= int64(endpoint.Weight)
	}

	return sqlc.TrunkEndpoint{}, fmt.Errorf("weighted carrier endpoint selection exhausted eligible endpoints")
}

func secureWeightPick(total int64) (int64, error) {
	if total <= 0 {
		return 0, fmt.Errorf("total endpoint weight must be positive")
	}
	pick, err := rand.Int(rand.Reader, big.NewInt(total))
	if err != nil {
		return 0, err
	}
	return pick.Int64(), nil
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
