package managedvoice

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/entitlements"
	"github.com/leamout/leamout/internal/telecom/calls"
)

const (
	FeatureEnabled        entitlements.Feature = "voice.managed.enabled"
	LimitDailySpendMicros                      = "voice.managed.max_daily_spend_micros"
)

type entitlementResolver interface {
	EffectiveForOrganization(context.Context, uuid.UUID) (entitlements.EntitlementSet, error)
}

type spendStore interface {
	ManagedSpendMicros(context.Context, uuid.UUID, time.Time) (int64, error)
}

type Policy struct {
	entitlements entitlementResolver
	spend        spendStore
	now          func() time.Time
}

func NewPolicy(resolver entitlementResolver, spend spendStore) (*Policy, error) {
	if resolver == nil || spend == nil {
		return nil, fmt.Errorf("managed voice entitlements and spend store are required")
	}
	return &Policy{entitlements: resolver, spend: spend, now: time.Now}, nil
}

func (p *Policy) AuthorizeManagedOutbound(ctx context.Context, organizationID uuid.UUID) error {
	set, err := p.entitlements.EffectiveForOrganization(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("resolve managed voice entitlements: %w", err)
	}
	if !set.Enabled(FeatureEnabled) {
		return calls.ErrManagedVoiceDisabled
	}
	limit, ok := set.Limit(LimitDailySpendMicros)
	if !ok || limit <= 0 {
		return calls.ErrManagedSpendLimit
	}
	now := p.now().UTC()
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	spend, err := p.spend.ManagedSpendMicros(ctx, organizationID, day)
	if err != nil {
		return fmt.Errorf("calculate managed voice spend: %w", err)
	}
	if spend >= limit {
		return calls.ErrManagedSpendLimit
	}
	return nil
}
