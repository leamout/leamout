package subscriptions

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/catalog"
)

type store interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (Subscription, error)
	Current(context.Context, uuid.UUID) (Subscription, error)
	List(context.Context, uuid.UUID) ([]Subscription, error)
	Create(context.Context, uuid.UUID, CreateInput) (Subscription, error)
	UpdatePlan(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (Subscription, error)
	UpdatePeriod(context.Context, uuid.UUID, uuid.UUID, PeriodUpdate) (Subscription, error)
	UpdateStatus(context.Context, uuid.UUID, uuid.UUID, Status, Status) (Subscription, error)
	SetProvider(context.Context, uuid.UUID, uuid.UUID, ProviderReference) (Subscription, error)
	GetByProvider(context.Context, ProviderReference) (Subscription, error)
}

type catalogReader interface {
	GetProduct(context.Context, uuid.UUID) (catalog.Product, error)
	GetPlan(context.Context, uuid.UUID) (catalog.Plan, error)
}

// Service owns subscription lifecycle and plan-eligibility rules.
type Service struct {
	store   store
	catalog catalogReader
	now     func() time.Time
}

func NewService(store store, catalog catalogReader) *Service {
	return &Service{store: store, catalog: catalog, now: time.Now}
}

func (s *Service) Create(ctx context.Context, organizationID uuid.UUID, input CreateInput) (Subscription, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return Subscription{}, err
	}
	normalized, err := normalizeCreate(input, s.now())
	if err != nil {
		return Subscription{}, err
	}
	if err := s.requireAvailablePlan(ctx, normalized.PlanID); err != nil {
		return Subscription{}, err
	}
	return s.store.Create(ctx, organizationID, normalized)
}

func (s *Service) Get(ctx context.Context, organizationID, id uuid.UUID) (Subscription, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return Subscription{}, err
	}
	if err := validateID(id, ErrSubscriptionIDRequired); err != nil {
		return Subscription{}, err
	}
	return s.store.Get(ctx, organizationID, id)
}

func (s *Service) Current(ctx context.Context, organizationID uuid.UUID) (Subscription, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return Subscription{}, err
	}
	return s.store.Current(ctx, organizationID)
}

func (s *Service) List(ctx context.Context, organizationID uuid.UUID) ([]Subscription, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return nil, err
	}
	return s.store.List(ctx, organizationID)
}

func (s *Service) ChangePlan(ctx context.Context, organizationID, id, planID uuid.UUID) (Subscription, error) {
	if err := validateID(planID, ErrPlanIDRequired); err != nil {
		return Subscription{}, err
	}
	current, err := s.Get(ctx, organizationID, id)
	if err != nil {
		return Subscription{}, err
	}
	if current.Status == StatusCancelled || current.Status == StatusExpired {
		return Subscription{}, ErrTerminalSubscription
	}
	if current.PlanID == planID {
		return current, nil
	}
	if err := s.requireAvailablePlan(ctx, planID); err != nil {
		return Subscription{}, err
	}
	return s.store.UpdatePlan(ctx, organizationID, id, planID)
}

func (s *Service) UpdatePeriod(ctx context.Context, organizationID, id uuid.UUID, input PeriodUpdate) (Subscription, error) {
	current, err := s.Get(ctx, organizationID, id)
	if err != nil {
		return Subscription{}, err
	}
	normalized, err := normalizePeriodUpdate(current, input)
	if err != nil {
		return Subscription{}, err
	}
	return s.store.UpdatePeriod(ctx, organizationID, id, normalized)
}

func (s *Service) Transition(ctx context.Context, organizationID, id uuid.UUID, to Status) (Subscription, error) {
	target, err := normalizeStatus(to)
	if err != nil {
		return Subscription{}, err
	}
	current, err := s.Get(ctx, organizationID, id)
	if err != nil {
		return Subscription{}, err
	}
	if err := validateTransition(current.Status, target); err != nil {
		return Subscription{}, err
	}
	if current.Status == target {
		return current, nil
	}
	return s.store.UpdateStatus(ctx, organizationID, id, current.Status, target)
}

func (s *Service) SetProvider(ctx context.Context, organizationID, id uuid.UUID, reference ProviderReference) (Subscription, error) {
	normalized, err := normalizeProvider(reference)
	if err != nil {
		return Subscription{}, err
	}
	current, err := s.Get(ctx, organizationID, id)
	if err != nil {
		return Subscription{}, err
	}
	if current.BillingProvider != nil && current.ProviderSubscriptionID != nil &&
		*current.BillingProvider == normalized.Provider && *current.ProviderSubscriptionID == normalized.SubscriptionID {
		return current, nil
	}
	return s.store.SetProvider(ctx, organizationID, id, normalized)
}

// GetByProvider resolves provider reconciliation metadata back to Leamout-owned state.
func (s *Service) GetByProvider(ctx context.Context, reference ProviderReference) (Subscription, error) {
	normalized, err := normalizeProvider(reference)
	if err != nil {
		return Subscription{}, err
	}
	return s.store.GetByProvider(ctx, normalized)
}

func (s *Service) requireAvailablePlan(ctx context.Context, planID uuid.UUID) error {
	plan, err := s.catalog.GetPlan(ctx, planID)
	if err != nil {
		if errors.Is(err, catalog.ErrPlanNotFound) {
			return ErrPlanUnavailable
		}
		return err
	}
	if !plan.Active {
		return ErrPlanUnavailable
	}
	product, err := s.catalog.GetProduct(ctx, plan.ProductID)
	if err != nil {
		if errors.Is(err, catalog.ErrProductNotFound) {
			return ErrPlanUnavailable
		}
		return err
	}
	if !product.Active {
		return ErrPlanUnavailable
	}
	return nil
}
