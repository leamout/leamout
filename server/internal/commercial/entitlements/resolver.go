package entitlements

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
)

var (
	ErrSourceRequired = errors.New("entitlement source is required")
	ErrKindConflict   = errors.New("entitlement kind conflicts with a lower-precedence value")
)

type ResolveRequest struct {
	OrganizationID uuid.UUID
	PlanID         uuid.UUID
	LicenseID      uuid.UUID
	At             time.Time
}

// Resolver combines plan defaults, organization overrides, and license
// overrides into a deterministic effective entitlement snapshot.
type Resolver struct {
	source Source
}

func NewResolver(source Source) (*Resolver, error) {
	if source == nil {
		return nil, ErrSourceRequired
	}
	return &Resolver{source: source}, nil
}

func (r *Resolver) Resolve(ctx context.Context, request ResolveRequest) ([]Entitlement, error) {
	if request.OrganizationID == uuid.Nil {
		return nil, fmt.Errorf("organization_id is required")
	}
	if request.At.IsZero() {
		return nil, fmt.Errorf("resolution time is required")
	}

	layers := make([][]Entitlement, 0, 3)
	if request.PlanID != uuid.Nil {
		plan, err := r.source.ListPlan(ctx, request.PlanID)
		if err != nil {
			return nil, fmt.Errorf("load plan entitlements: %w", err)
		}
		layers = append(layers, plan)
	}
	organization, err := r.source.ListOrganization(ctx, request.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("load organization entitlements: %w", err)
	}
	layers = append(layers, organization)
	if request.LicenseID != uuid.Nil {
		license, err := r.source.ListLicense(ctx, request.OrganizationID, request.LicenseID)
		if err != nil {
			return nil, fmt.Errorf("load license entitlements: %w", err)
		}
		layers = append(layers, license)
	}

	effective := make(map[string]Entitlement)
	for _, layer := range layers {
		seen := make(map[string]struct{}, len(layer))
		for _, entitlement := range layer {
			if err := Validate(entitlement); err != nil {
				return nil, fmt.Errorf("invalid entitlement %q: %w", entitlement.Key, err)
			}
			if !entitlement.ActiveAt(request.At) {
				continue
			}
			if _, duplicate := seen[entitlement.Key]; duplicate {
				return nil, fmt.Errorf("duplicate entitlement %q in one layer", entitlement.Key)
			}
			seen[entitlement.Key] = struct{}{}
			if current, exists := effective[entitlement.Key]; exists && current.Kind != entitlement.Kind {
				return nil, fmt.Errorf("%w: %q is %q and %q", ErrKindConflict, entitlement.Key, current.Kind, entitlement.Kind)
			}
			effective[entitlement.Key] = entitlement
		}
	}

	keys := make([]string, 0, len(effective))
	for key := range effective {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]Entitlement, 0, len(keys))
	for _, key := range keys {
		result = append(result, effective[key])
	}
	return result, nil
}
