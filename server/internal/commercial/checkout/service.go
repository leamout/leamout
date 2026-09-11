package checkout

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type store interface {
	Create(context.Context, uuid.UUID, CreateInput) (Checkout, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (Checkout, error)
	GetByReference(context.Context, string) (Checkout, error)
	Transition(context.Context, uuid.UUID, uuid.UUID, Transition) (Checkout, error)
	ClaimRefresh(context.Context, uuid.UUID, uuid.UUID, time.Time) (Checkout, error)
	Expire(context.Context) ([]Checkout, error)
}

// Service owns checkout validation and lifecycle use cases.
type Service struct {
	repo store
	now  func() time.Time
}

func NewService(repo store) *Service { return &Service{repo: repo, now: time.Now} }

func (s *Service) Create(ctx context.Context, organizationID uuid.UUID, input CreateInput) (Checkout, error) {
	if organizationID == uuid.Nil || validateCreate(input, s.now()) != nil {
		return Checkout{}, ErrInvalidCheckout
	}
	return s.repo.Create(ctx, organizationID, input)
}
func (s *Service) Get(ctx context.Context, organizationID, id uuid.UUID) (Checkout, error) {
	return s.repo.Get(ctx, organizationID, id)
}
func (s *Service) GetByReference(ctx context.Context, reference string) (Checkout, error) {
	return s.repo.GetByReference(ctx, reference)
}
func (s *Service) Transition(ctx context.Context, organizationID, id uuid.UUID, transition Transition) (Checkout, error) {
	if validateTransition(transition) != nil {
		return Checkout{}, ErrInvalidTransition
	}
	return s.repo.Transition(ctx, organizationID, id, transition)
}
func (s *Service) ClaimRefresh(ctx context.Context, organizationID, id uuid.UUID, before time.Time) (Checkout, error) {
	return s.repo.ClaimRefresh(ctx, organizationID, id, before)
}
func (s *Service) Expire(ctx context.Context) ([]Checkout, error) { return s.repo.Expire(ctx) }
