package licensing

import "context"

// Service owns commercial license lifecycle operations. Signing and activation
// details are intentionally kept behind dedicated collaborators.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, id string) (License, error) {
	return s.repo.Get(ctx, id)
}
