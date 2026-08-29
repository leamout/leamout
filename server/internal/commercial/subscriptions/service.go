package subscriptions

import "context"

// Service owns subscription lifecycle operations independently of payment providers.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, id string) (Subscription, error) {
	return s.repo.Get(ctx, id)
}
