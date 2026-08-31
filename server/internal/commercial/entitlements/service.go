package entitlements

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
)

type store interface {
	CreatePlan(context.Context, uuid.UUID, CreateInput) (Entitlement, error)
	CreateOrganization(context.Context, uuid.UUID, CreateInput) (Entitlement, error)
	CreateLicense(context.Context, uuid.UUID, uuid.UUID, CreateInput) (Entitlement, error)
	ListPlan(context.Context, uuid.UUID) ([]Entitlement, error)
	ListOrganization(context.Context, uuid.UUID) ([]Entitlement, error)
	ListLicense(context.Context, uuid.UUID, uuid.UUID) ([]Entitlement, error)
	DeletePlan(context.Context, uuid.UUID, uuid.UUID) error
	DeleteOrganization(context.Context, uuid.UUID, uuid.UUID) error
	DeleteLicense(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) error
}

type subscriptionReader interface {
	Current(context.Context, uuid.UUID) (subscriptions.Subscription, error)
}

// Service owns entitlement validation, scope operations, and effective resolution.
type Service struct {
	store         store
	subscriptions subscriptionReader
	now           func() time.Time
}

func NewService(store store, subscriptions subscriptionReader) *Service {
	return &Service{store: store, subscriptions: subscriptions, now: time.Now}
}

func (s *Service) CreatePlan(ctx context.Context, planID uuid.UUID, input CreateInput) (Entitlement, error) {
	if err := validateID(planID, ErrPlanIDRequired); err != nil {
		return Entitlement{}, err
	}
	normalized, err := normalizeCreate(input)
	if err != nil {
		return Entitlement{}, err
	}
	return s.store.CreatePlan(ctx, planID, normalized)
}

func (s *Service) CreateOrganization(ctx context.Context, organizationID uuid.UUID, input CreateInput) (Entitlement, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return Entitlement{}, err
	}
	normalized, err := normalizeCreate(input)
	if err != nil {
		return Entitlement{}, err
	}
	return s.store.CreateOrganization(ctx, organizationID, normalized)
}

func (s *Service) CreateLicense(ctx context.Context, organizationID, licenseID uuid.UUID, input CreateInput) (Entitlement, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return Entitlement{}, err
	}
	if err := validateID(licenseID, ErrLicenseIDRequired); err != nil {
		return Entitlement{}, err
	}
	normalized, err := normalizeCreate(input)
	if err != nil {
		return Entitlement{}, err
	}
	return s.store.CreateLicense(ctx, organizationID, licenseID, normalized)
}

func (s *Service) ListPlan(ctx context.Context, planID uuid.UUID) ([]Entitlement, error) {
	if err := validateID(planID, ErrPlanIDRequired); err != nil {
		return nil, err
	}
	return s.store.ListPlan(ctx, planID)
}

func (s *Service) ListOrganization(ctx context.Context, organizationID uuid.UUID) ([]Entitlement, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return nil, err
	}
	return s.store.ListOrganization(ctx, organizationID)
}

func (s *Service) ListLicense(ctx context.Context, organizationID, licenseID uuid.UUID) ([]Entitlement, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return nil, err
	}
	if err := validateID(licenseID, ErrLicenseIDRequired); err != nil {
		return nil, err
	}
	return s.store.ListLicense(ctx, organizationID, licenseID)
}

func (s *Service) DeletePlan(ctx context.Context, planID, id uuid.UUID) error {
	if err := validateID(planID, ErrPlanIDRequired); err != nil {
		return err
	}
	if err := validateID(id, ErrEntitlementIDRequired); err != nil {
		return err
	}
	return s.store.DeletePlan(ctx, planID, id)
}

func (s *Service) DeleteOrganization(ctx context.Context, organizationID, id uuid.UUID) error {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return err
	}
	if err := validateID(id, ErrEntitlementIDRequired); err != nil {
		return err
	}
	return s.store.DeleteOrganization(ctx, organizationID, id)
}

func (s *Service) DeleteLicense(ctx context.Context, organizationID, licenseID, id uuid.UUID) error {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return err
	}
	if err := validateID(licenseID, ErrLicenseIDRequired); err != nil {
		return err
	}
	if err := validateID(id, ErrEntitlementIDRequired); err != nil {
		return err
	}
	return s.store.DeleteLicense(ctx, organizationID, licenseID, id)
}

// EffectiveForOrganization resolves plan defaults and organization overrides at the current time.
func (s *Service) EffectiveForOrganization(ctx context.Context, organizationID uuid.UUID) (EntitlementSet, error) {
	return s.EffectiveForOrganizationAt(ctx, organizationID, s.now())
}

// EffectiveForOrganizationAt resolves organization capabilities at a deterministic evaluation time.
func (s *Service) EffectiveForOrganizationAt(ctx context.Context, organizationID uuid.UUID, at time.Time) (EntitlementSet, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return EntitlementSet{}, err
	}
	current, err := s.subscriptions.Current(ctx, organizationID)
	if err != nil {
		if errors.Is(err, subscriptions.ErrSubscriptionNotFound) {
			return EntitlementSet{}, ErrSubscriptionUnavailable
		}
		return EntitlementSet{}, err
	}
	return s.EffectiveForOrganizationPlanAt(ctx, organizationID, current.PlanID, at)
}

// EffectiveForOrganizationPlanAt resolves organization capabilities against an already-selected plan.
// Callers that already resolved the current subscription can use this method to keep one state read coherent.
func (s *Service) EffectiveForOrganizationPlanAt(ctx context.Context, organizationID, planID uuid.UUID, at time.Time) (EntitlementSet, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return EntitlementSet{}, err
	}
	if err := validateID(planID, ErrPlanIDRequired); err != nil {
		return EntitlementSet{}, err
	}
	plan, err := s.store.ListPlan(ctx, planID)
	if err != nil {
		return EntitlementSet{}, err
	}
	overrides, err := s.store.ListOrganization(ctx, organizationID)
	if err != nil {
		return EntitlementSet{}, err
	}
	return resolve(at, plan, overrides)
}

// ForLicense resolves the immutable license-scoped entitlement snapshot at the current time.
func (s *Service) ForLicense(ctx context.Context, organizationID, licenseID uuid.UUID) (EntitlementSet, error) {
	return s.ForLicenseAt(ctx, organizationID, licenseID, s.now())
}

// ForLicenseAt resolves a license snapshot at a deterministic evaluation time.
func (s *Service) ForLicenseAt(ctx context.Context, organizationID, licenseID uuid.UUID, at time.Time) (EntitlementSet, error) {
	rows, err := s.ListLicense(ctx, organizationID, licenseID)
	if err != nil {
		return EntitlementSet{}, err
	}
	return resolve(at, rows, nil)
}

func resolve(at time.Time, base, overrides []Entitlement) (EntitlementSet, error) {
	resolved := make(map[string]Entitlement, len(base)+len(overrides))
	for _, entitlement := range base {
		if effectiveAt(entitlement, at) {
			resolved[entitlement.Key] = entitlement
		}
	}
	for _, entitlement := range overrides {
		if !effectiveAt(entitlement, at) {
			continue
		}
		if inherited, ok := resolved[entitlement.Key]; ok && inherited.Kind != entitlement.Kind {
			return EntitlementSet{}, ErrKindMismatch
		}
		resolved[entitlement.Key] = entitlement
	}

	set := EntitlementSet{
		Features: make(map[Feature]bool),
		Limits:   make(map[string]int64),
	}
	for key, entitlement := range resolved {
		switch entitlement.Kind {
		case KindFeature:
			if entitlement.Enabled == nil {
				return EntitlementSet{}, ErrInvalidEntitlement
			}
			set.Features[Feature(key)] = *entitlement.Enabled
		case KindLimit:
			if entitlement.LimitValue == nil || *entitlement.LimitValue < 0 {
				return EntitlementSet{}, ErrInvalidEntitlement
			}
			set.Limits[key] = *entitlement.LimitValue
		default:
			return EntitlementSet{}, ErrInvalidEntitlement
		}
	}
	return set, nil
}
