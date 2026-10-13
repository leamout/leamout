package subscriptions

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/server/pkg/apperror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Get(ctx context.Context, organizationID uuid.UUID) (*Subscription, error) {
	if organizationID == uuid.Nil {
		return nil, apperror.NewBadRequest("organization_id is required")
	}
	value, err := s.repo.Get(ctx, organizationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, apperror.NewInternal("get organization subscription", err)
	}
	return &value, nil
}

func (s *Service) Sync(
	ctx context.Context,
	organizationID uuid.UUID,
	req SyncRequest,
) (Subscription, error) {
	if organizationID == uuid.Nil {
		return Subscription{}, apperror.NewBadRequest("organization_id is required")
	}
	if err := validateSync(req); err != nil {
		return Subscription{}, err
	}
	existing, err := s.Get(ctx, organizationID)
	if err != nil {
		return Subscription{}, err
	}
	id := uuid.New()
	if existing != nil {
		id = existing.ID
	}
	value, err := s.repo.Upsert(ctx, id, organizationID, req)
	if errors.Is(err, pgx.ErrNoRows) {
		// Out-of-order or conflicting provider events are no-ops.
		current, readErr := s.Get(ctx, organizationID)
		if readErr != nil {
			return Subscription{}, readErr
		}
		if current != nil {
			return *current, nil
		}
		return Subscription{}, apperror.NewServiceUnavailable("subscription synchronization race", err)
	}
	if err != nil {
		return Subscription{}, apperror.NewInternal("sync organization subscription", err)
	}
	return value, nil
}

// PaidAccessActive applies Leamout's payment policy independently of Polar's
// benefit grace period. Past-due access ends at the persisted grace deadline,
// including when the provider's last paid period has already ended.
func PaidAccessActive(sub *Subscription, now time.Time) bool {
	if sub == nil || sub.CurrentPeriodStart == nil || sub.CurrentPeriodEnd == nil {
		return false
	}
	if now.Before(*sub.CurrentPeriodStart) || !sub.CurrentPeriodEnd.After(*sub.CurrentPeriodStart) {
		return false
	}
	switch sub.Status {
	case StatusActive:
		return now.Before(*sub.CurrentPeriodEnd)
	case StatusTrialing:
		return now.Before(*sub.CurrentPeriodEnd) &&
			(sub.TrialEndsAt == nil || now.Before(*sub.TrialEndsAt))
	case StatusPastDue:
		return sub.PastDueAt != nil && !now.Before(*sub.PastDueAt) &&
			sub.GraceEndsAt != nil && now.Before(*sub.GraceEndsAt)
	default:
		return false
	}
}
