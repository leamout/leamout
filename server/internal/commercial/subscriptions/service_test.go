package subscriptions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/catalog"
)

type fakeStore struct {
	current  Subscription
	created  CreateInput
	expected Status
	status   Status
	planID   uuid.UUID
}

func (f *fakeStore) Get(context.Context, uuid.UUID, uuid.UUID) (Subscription, error) {
	return f.current, nil
}
func (f *fakeStore) Current(context.Context, uuid.UUID) (Subscription, error) { return f.current, nil }
func (f *fakeStore) List(context.Context, uuid.UUID) ([]Subscription, error) {
	return []Subscription{f.current}, nil
}
func (f *fakeStore) Create(_ context.Context, organizationID uuid.UUID, input CreateInput) (Subscription, error) {
	f.created = input
	return Subscription{ID: uuid.New(), OrganizationID: organizationID, PlanID: input.PlanID, Status: StatusPending, StartsAt: *input.StartsAt}, nil
}
func (f *fakeStore) UpdatePlan(_ context.Context, _, _ uuid.UUID, planID uuid.UUID) (Subscription, error) {
	f.planID = planID
	f.current.PlanID = planID
	return f.current, nil
}
func (f *fakeStore) UpdatePeriod(_ context.Context, _, _ uuid.UUID, input PeriodUpdate) (Subscription, error) {
	f.current.RenewsAt = input.RenewsAt
	f.current.EndsAt = input.EndsAt
	return f.current, nil
}
func (f *fakeStore) UpdateStatus(_ context.Context, _, _ uuid.UUID, expected, status Status) (Subscription, error) {
	f.expected = expected
	f.status = status
	f.current.Status = status
	return f.current, nil
}
func (f *fakeStore) SetProvider(_ context.Context, _, _ uuid.UUID, reference ProviderReference) (Subscription, error) {
	f.current.BillingProvider = &reference.Provider
	f.current.ProviderSubscriptionID = &reference.SubscriptionID
	return f.current, nil
}
func (f *fakeStore) GetByProvider(context.Context, ProviderReference) (Subscription, error) {
	return f.current, nil
}

type fakeCatalog struct {
	product catalog.Product
	plan    catalog.Plan
}

func (f fakeCatalog) GetProduct(context.Context, uuid.UUID) (catalog.Product, error) {
	if f.product.ID == uuid.Nil {
		return catalog.Product{}, catalog.ErrProductNotFound
	}
	return f.product, nil
}
func (f fakeCatalog) GetPlan(context.Context, uuid.UUID) (catalog.Plan, error) {
	if f.plan.ID == uuid.Nil {
		return catalog.Plan{}, catalog.ErrPlanNotFound
	}
	return f.plan, nil
}

func TestCreateRequiresAvailablePlan(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	planID := uuid.New()
	productID := uuid.New()
	store := &fakeStore{}
	service := NewService(store, fakeCatalog{
		product: catalog.Product{ID: productID, Active: true},
		plan:    catalog.Plan{ID: planID, ProductID: productID, Active: false},
	})

	_, err := service.Create(context.Background(), orgID, CreateInput{PlanID: planID})
	if !errors.Is(err, ErrPlanUnavailable) {
		t.Fatalf("expected ErrPlanUnavailable, got %v", err)
	}
}

func TestCreateNormalizesStartTime(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	planID := uuid.New()
	productID := uuid.New()
	now := time.Date(2026, 8, 31, 4, 0, 0, 0, time.UTC)
	store := &fakeStore{}
	service := NewService(store, fakeCatalog{
		product: catalog.Product{ID: productID, Active: true},
		plan:    catalog.Plan{ID: planID, ProductID: productID, Active: true},
	})
	service.now = func() time.Time { return now }

	_, err := service.Create(context.Background(), orgID, CreateInput{PlanID: planID})
	if err != nil {
		t.Fatalf("create subscription: %v", err)
	}
	if store.created.StartsAt == nil || !store.created.StartsAt.Equal(now) {
		t.Fatalf("expected normalized starts_at %v, got %#v", now, store.created.StartsAt)
	}
}

func TestTransitionIsIdempotent(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	id := uuid.New()
	store := &fakeStore{current: Subscription{ID: id, OrganizationID: orgID, Status: StatusActive}}
	service := NewService(store, fakeCatalog{})

	got, err := service.Transition(context.Background(), orgID, id, StatusActive)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if got.Status != StatusActive {
		t.Fatalf("expected active status, got %s", got.Status)
	}
	if store.status != "" {
		t.Fatalf("expected no persistence write for idempotent transition, got %s", store.status)
	}
}

func TestTransitionUsesExpectedStatus(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	id := uuid.New()
	store := &fakeStore{current: Subscription{ID: id, OrganizationID: orgID, Status: StatusActive}}
	service := NewService(store, fakeCatalog{})

	_, err := service.Transition(context.Background(), orgID, id, StatusPastDue)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if store.expected != StatusActive || store.status != StatusPastDue {
		t.Fatalf("expected active -> past_due compare-and-set, got %s -> %s", store.expected, store.status)
	}
}

func TestChangePlanRejectsTerminalSubscription(t *testing.T) {
	t.Parallel()

	orgID := uuid.New()
	id := uuid.New()
	store := &fakeStore{current: Subscription{ID: id, OrganizationID: orgID, Status: StatusCancelled}}
	service := NewService(store, fakeCatalog{})

	_, err := service.ChangePlan(context.Background(), orgID, id, uuid.New())
	if !errors.Is(err, ErrTerminalSubscription) {
		t.Fatalf("expected ErrTerminalSubscription, got %v", err)
	}
}
