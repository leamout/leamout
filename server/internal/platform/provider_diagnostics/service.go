package provider_diagnostics

import (
	"context"
	"errors"
)

type store interface {
	Snapshot(context.Context, int32) (Snapshot, error)
}

type Service struct {
	store store
}

func NewService(store store) *Service {
	return &Service{store: store}
}

func (s *Service) Snapshot(ctx context.Context, limit int32) (Snapshot, error) {
	if s == nil || s.store == nil {
		return Snapshot{}, errors.New("provider diagnostics unavailable")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	return s.store.Snapshot(ctx, limit)
}
