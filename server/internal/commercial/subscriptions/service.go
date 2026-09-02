package subscriptions

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/catalog"
)

// Service owns subscription lifecycle and acquired commercial terms.
type Service struct {
	repo    *Repository
	catalog *catalog.Service
	now     func() time.Time
}

func NewService(repo *Repository, catalog *catalog.Service) *Service {
	return &Service{repo: repo, catalog: catalog, now: time.Now}
}

func (s *Service) Create(ctx context.Context, organizationID uuid.UUID, input CreateInput) (Subscription, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return Subscription{}, err
	}
	normalized, err := normalizeCreate(input, s.now())
	if err != nil {
		return Subscription{}, err
	}
	price, err := s.requireAvailablePrice(ctx, normalized.PriceID, *normalized.StartsAt)
	if err != nil {
		return Subscription{}, err
	}
	return s.repo.Create(ctx, organizationID, price.PlanID, normalized)
}

func (s *Service) Get(ctx context.Context, organizationID, id uuid.UUID) (Subscription, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return Subscription{}, err
	}
	if err := validateID(id, ErrSubscriptionIDRequired); err != nil {
		return Subscription{}, err
	}
	return s.repo.Get(ctx, organizationID, id)
}

func (s *Service) Current(ctx context.Context, organizationID uuid.UUID) (Subscription, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return Subscription{}, err
	}
	return s.repo.Current(ctx, organizationID)
}

func (s *Service) List(ctx context.Context, organizationID uuid.UUID) ([]Subscription, error) {
	if err := validateID(organizationID, ErrOrganizationIDRequired); err != nil {
		return nil, err
	}
	return s.repo.List(ctx, organizationID)
}

// ChangePrice changes the subscription's acquired commercial terms. The selected
// price determines the new plan and keeps price/plan identity consistent.
func (s *Service) ChangePrice(ctx context.Context, organizationID, id, priceID uuid.UUID) (Subscription, error) {
	if err := validateID(priceID, ErrPriceIDRequired); err != nil {
		return Subscription{}, err
	}
	current, err := s.Get(ctx, organizationID, id)
	if err != nil {
		return Subscription{}, err
	}
	if current.Status == StatusCancelled || current.Status == StatusExpired {
		return Subscription{}, ErrTerminalSubscription
	}
	if current.PriceID != nil && *current.PriceID == priceID {
		return current, nil
	}
	price, err := s.requireAvailablePrice(ctx, priceID, s.now())
	if err != nil {
		return Subscription{}, err
	}
	return s.repo.UpdatePrice(ctx, organizationID, id, price.ID, price.PlanID)
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
	return s.repo.UpdatePeriod(ctx, organizationID, id, normalized)
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
	return s.repo.UpdateStatus(ctx, organizationID, id, current.Status, target)
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
	return s.repo.SetProvider(ctx, organizationID, id, normalized)
}

// GetByProvider resolves provider reconciliation metadata back to Leamout-owned state.
func (s *Service) GetByProvider(ctx context.Context, reference ProviderReference) (Subscription, error) {
	normalized, err := normalizeProvider(reference)
	if err != nil {
		return Subscription{}, err
	}
	return s.repo.GetByProvider(ctx, normalized)
}

func (s *Service) requireAvailablePrice(ctx context.Context, priceID uuid.UUID, at time.Time) (catalog.Price, error) {
	price, err := s.catalog.GetPrice(ctx, priceID)
	if err != nil {
		if errors.Is(err, catalog.ErrPriceNotFound) {
			return catalog.Price{}, ErrPriceUnavailable
		}
		return catalog.Price{}, err
	}
	if !price.EffectiveAt(at) {
		return catalog.Price{}, ErrPriceUnavailable
	}
	plan, err := s.catalog.GetPlan(ctx, price.PlanID)
	if err != nil {
		if errors.Is(err, catalog.ErrPlanNotFound) {
			return catalog.Price{}, ErrPriceUnavailable
		}
		return catalog.Price{}, err
	}
	if !plan.Active {
		return catalog.Price{}, ErrPriceUnavailable
	}
	product, err := s.catalog.GetProduct(ctx, plan.ProductID)
	if err != nil {
		if errors.Is(err, catalog.ErrProductNotFound) {
			return catalog.Price{}, ErrPriceUnavailable
		}
		return catalog.Price{}, err
	}
	if !product.Active {
		return catalog.Price{}, ErrPriceUnavailable
	}
	return price, nil
}
