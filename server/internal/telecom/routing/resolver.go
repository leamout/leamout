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
	ListManagedOutboundRoutes(context.Context) ([]managedRouteCandidate, error)
	ResolveInboundCarrier(context.Context, netip.Addr) (sqlc.CarrierConnection, error)
	GetPhoneNumber(context.Context, uuid.UUID, string) (sqlc.PhoneNumber, error)
	ResolveInboundPhoneNumber(context.Context, uuid.UUID, string) (sqlc.PhoneNumber, error)
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
	if req.TrunkID != nil {
		return r.resolveBYOCOutbound(ctx, req, *req.TrunkID)
	}
	return r.resolveManagedOutbound(ctx, req)
}

func (r *Resolver) resolveBYOCOutbound(
	ctx context.Context,
	req OutboundRequest,
	trunkID uuid.UUID,
) (OutboundDecision, error) {
	trunk, err := r.repo.GetTrunk(ctx, req.OrganizationID, trunkID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OutboundDecision{}, ErrNoRoute
		}
		return OutboundDecision{}, err
	}
	if trunk.OrganizationID == nil || *trunk.OrganizationID != req.OrganizationID {
		return OutboundDecision{}, ErrNoRoute
	}

	connection, err := r.repo.GetCarrierConnection(ctx, req.OrganizationID, trunk.CarrierConnectionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OutboundDecision{}, ErrNoRoute
		}
		return OutboundDecision{}, err
	}
	if connection.Scope != "organization" || connection.OrganizationID == nil || *connection.OrganizationID != req.OrganizationID {
		return OutboundDecision{}, ErrNoRoute
	}

	caller, err := r.repo.GetPhoneNumber(ctx, req.OrganizationID, req.From)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OutboundDecision{}, ErrCallerIdentity
		}
		return OutboundDecision{}, err
	}
	if err := authorizeBYOCCallerIdentity(caller, req.OrganizationID, trunk.CarrierConnectionID); err != nil {
		return OutboundDecision{}, err
	}

	endpoints, err := r.repo.ListOutboundEndpoints(ctx, req.OrganizationID, trunkID)
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
	r.recordEndpointSelection(ctx, trunk.CarrierConnectionID, trunk.ID, endpoint, endpoints)
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

func (r *Resolver) resolveManagedOutbound(
	ctx context.Context,
	req OutboundRequest,
) (OutboundDecision, error) {
	caller, err := r.repo.GetPhoneNumber(ctx, req.OrganizationID, req.From)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return OutboundDecision{}, ErrCallerIdentity
		}
		return OutboundDecision{}, err
	}
	if err := authorizeManagedCallerIdentity(caller, req.OrganizationID); err != nil {
		return OutboundDecision{}, err
	}

	candidates, err := r.repo.ListManagedOutboundRoutes(ctx)
	if err != nil {
		return OutboundDecision{}, err
	}
	if len(candidates) == 0 {
		return OutboundDecision{}, ErrNoRoute
	}

	endpoints := make([]sqlc.TrunkEndpoint, 0, len(candidates))
	for _, candidate := range candidates {
		endpoints = append(endpoints, sqlc.TrunkEndpoint{
			ID:           candidate.EndpointID,
			TrunkID:      candidate.TrunkID,
			Host:         candidate.Host,
			Port:         candidate.Port,
			Transport:    candidate.Transport,
			Priority:     candidate.Priority,
			Weight:       candidate.Weight,
			HealthStatus: candidate.HealthStatus,
		})
	}

	endpoint, err := r.selectOutboundEndpoint(endpoints)
	if err != nil {
		return OutboundDecision{}, err
	}

	var selected *managedRouteCandidate
	for i := range candidates {
		if candidates[i].EndpointID == endpoint.ID {
			selected = &candidates[i]
			break
		}
	}
	if selected == nil {
		return OutboundDecision{}, ErrNoRoute
	}

	r.recordEndpointSelection(ctx, selected.CarrierConnectionID, selected.TrunkID, endpoint, endpoints)
	return OutboundDecision{
		Managed:             true,
		OrganizationID:      req.OrganizationID,
		CarrierConnectionID: selected.CarrierConnectionID,
		TrunkID:             selected.TrunkID,
		EndpointID:          selected.EndpointID,
		Host:                selected.Host,
		Port:                selected.Port,
		Transport:           selected.Transport,
		From:                req.From,
		To:                  req.To,
		MaxCPS:              selected.MaxCPS,
		MaxConcurrentCalls:  selected.MaxConcurrentCalls,
		MaxDailyMinutes:     selected.MaxDailyMinutes,
	}, nil
}

func (r *Resolver) recordEndpointSelection(
	ctx context.Context,
	connectionID uuid.UUID,
	trunkID uuid.UUID,
	endpoint sqlc.TrunkEndpoint,
	endpoints []sqlc.TrunkEndpoint,
) {
	if r.metrics == nil {
		return
	}
	failover := len(endpoints) > 0 && endpoint.Priority > endpoints[0].Priority
	r.metrics.EndpointSelection(ctx, connectionID, trunkID, endpoint.ID, failover)
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

	phoneNumber, err := r.repo.ResolveInboundPhoneNumber(ctx, connection.ID, req.CalledNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InboundDecision{}, ErrNoRoute
		}
		return InboundDecision{}, err
	}
	organizationID := phoneNumber.OrganizationID

	switch connection.Scope {
	case "organization":
		if connection.OrganizationID == nil ||
			*connection.OrganizationID != organizationID ||
			phoneNumber.ProvisioningMode != provisioningModeBYOC {
			return InboundDecision{}, ErrTenantMismatch
		}
	case "platform":
		if connection.OrganizationID != nil || phoneNumber.ProvisioningMode != provisioningModeManaged {
			return InboundDecision{}, ErrTenantMismatch
		}
	default:
		return InboundDecision{}, ErrTenantMismatch
	}

	binding, err := r.repo.GetVoiceBinding(ctx, req.CalledNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InboundDecision{}, ErrNoRoute
		}
		return InboundDecision{}, err
	}
	if binding.OrganizationID != organizationID {
		return InboundDecision{}, ErrTenantMismatch
	}

	return InboundDecision{
		OrganizationID:      organizationID,
		CarrierConnectionID: connection.ID,
		PhoneNumberID:       phoneNumber.ID,
		VoiceApplicationID:  binding.VoiceApplicationID,
		CalledNumber:        req.CalledNumber,
		CallerNumber:        req.CallerNumber,
	}, nil
}
