package limits

import (
	"context"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/commercial/plans"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
	"github.com/leamout/leamout/server/pkg/apperror"
)

type Service struct {
	plans         *plans.Service
	subscriptions *subscriptions.Service
}

func NewService(plansService *plans.Service, subscriptionsService *subscriptions.Service) *Service {
	return &Service{
		plans:         plansService,
		subscriptions: subscriptionsService,
	}
}

func (s *Service) Get(ctx context.Context, organizationID uuid.UUID) (Effective, error) {
	if organizationID == uuid.Nil {
		return Effective{}, apperror.NewBadRequest("organization_id is required")
	}

	subscription, err := s.subscriptions.Get(ctx, organizationID)
	if err != nil {
		return Effective{}, err
	}

	var plan plans.Plan
	if subscription == nil ||
		(subscription.Status != subscriptions.StatusActive &&
			subscription.Status != subscriptions.StatusTrialing) {
		plan, err = s.plans.GetByCode(ctx, "free")
	} else {
		plan, err = s.plans.GetByID(ctx, subscription.PlanID)
	}
	if err != nil {
		return Effective{}, err
	}

	return fromPlan(plan), nil
}

func (s *Service) MaxConcurrentCalls(
	ctx context.Context,
	organizationID uuid.UUID,
) (int64, bool, error) {
	value, err := s.Get(ctx, organizationID)
	if err != nil {
		return 0, false, err
	}
	if value.MaxConcurrentCalls == nil {
		return 0, false, nil
	}
	return *value.MaxConcurrentCalls, true, nil
}

func (s *Service) RetentionDays(
	ctx context.Context,
	organizationID uuid.UUID,
) (int64, bool, error) {
	value, err := s.Get(ctx, organizationID)
	if err != nil {
		return 0, false, err
	}
	if value.RetentionDays == nil {
		return 0, false, nil
	}
	return *value.RetentionDays, true, nil
}

func fromPlan(plan plans.Plan) Effective {
	return Effective{
		MaxConcurrentCalls: positiveLimit(plan.Limits[plans.LimitMaxConcurrentCalls]),
		RetentionDays:      positiveLimit(plan.Limits[plans.LimitRetentionDays]),
	}
}

func positiveLimit(value int64) *int64 {
	if value < 1 {
		return nil
	}
	return &value
}
