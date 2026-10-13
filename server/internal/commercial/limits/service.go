package limits

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/commercial/licensing"
	"github.com/leamout/leamout/server/internal/commercial/plans"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/pkg/apperror"
)

type Service struct {
	license       *licensing.Service
	plans         *plans.Service
	subscriptions *subscriptions.Service
	queries       *sqlc.Queries
	now           func() time.Time
}

func NewService(plansService *plans.Service, subscriptionsService *subscriptions.Service, queries ...*sqlc.Queries) *Service {
	s := &Service{
		plans:         plansService,
		subscriptions: subscriptionsService,
		now:           time.Now,
	}
	if len(queries) > 0 {
		s.queries = queries[0]
	}
	return s
}

func (s *Service) SetLicenseService(service *licensing.Service) {
	s.license = service
}

func (s *Service) Get(ctx context.Context, organizationID uuid.UUID) (Effective, error) {
	if organizationID == uuid.Nil {
		return Effective{}, apperror.NewBadRequest("organization_id is required")
	}
	if s.license.SelfHosted() {
		claims, err := s.license.RequireRuntime()
		if err != nil {
			return Effective{}, err
		}
		return Effective{
			MaxConcurrentCalls: &claims.MaxConcurrentCalls,
			PeriodStart:        claims.ValidFrom,
			PeriodEnd:          claims.ExpiresAt,
		}, nil
	}
	sub, err := s.subscriptions.Get(ctx, organizationID)
	if err != nil {
		return Effective{}, err
	}
	now := s.now().UTC()
	start, end := calendarMonth(now)
	valid := subscriptionPeriodActive(sub, now)
	var plan plans.Plan
	if valid {
		plan, err = s.plans.GetByID(ctx, sub.PlanID)
		start, end = sub.CurrentPeriodStart.UTC(), sub.CurrentPeriodEnd.UTC()
	} else {
		plan, err = s.plans.GetByCode(ctx, "free")
	}
	if err != nil {
		return Effective{}, err
	}
	result := fromPlan(plan)
	result.PeriodStart = start
	result.PeriodEnd = end
	if !validPlanLimits(result) {
		return Effective{}, apperror.NewServiceUnavailable("organization plan capacity is incomplete", nil)
	}
	return result, nil
}

func subscriptionPeriodActive(sub *subscriptions.Subscription, now time.Time) bool {
	if sub == nil || (sub.Status != subscriptions.StatusActive && sub.Status != subscriptions.StatusTrialing) {
		return false
	}
	if sub.CurrentPeriodStart == nil || sub.CurrentPeriodEnd == nil || now.Before(*sub.CurrentPeriodStart) || !now.Before(*sub.CurrentPeriodEnd) {
		return false
	}
	if sub.Status == subscriptions.StatusTrialing && sub.TrialEndsAt != nil && !now.Before(*sub.TrialEndsAt) {
		return false
	}
	return true
}

func calendarMonth(now time.Time) (time.Time, time.Time) {
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	return start, start.AddDate(0, 1, 0)
}

func validPlanLimits(v Effective) bool {
	return v.MaxConcurrentCalls != nil && v.RetentionDays != nil && v.MonthlyRuntimeMinutes != nil && v.MaxVoiceAgents != nil && v.MaxSIPTrunks != nil
}

func (s *Service) MaxConcurrentCalls(ctx context.Context, organizationID uuid.UUID) (int64, bool, error) {
	value, err := s.Get(ctx, organizationID)
	if err != nil {
		return 0, false, err
	}
	return *value.MaxConcurrentCalls, true, nil
}

func (s *Service) RetentionDays(ctx context.Context, organizationID uuid.UUID) (int64, bool, error) {
	value, err := s.Get(ctx, organizationID)
	if err != nil {
		return 0, false, err
	}
	if value.RetentionDays == nil {
		return 0, false, nil
	}
	return *value.RetentionDays, true, nil
}

func (s *Service) MaxVoiceAgents(ctx context.Context, organizationID uuid.UUID) (int64, bool, error) {
	value, err := s.Get(ctx, organizationID)
	if err != nil {
		return 0, false, err
	}
	if value.MaxVoiceAgents == nil {
		return 0, false, nil
	}
	return *value.MaxVoiceAgents, true, nil
}

func (s *Service) MaxSIPTrunks(ctx context.Context, organizationID uuid.UUID) (int64, bool, error) {
	value, err := s.Get(ctx, organizationID)
	if err != nil {
		return 0, false, err
	}
	if value.MaxSIPTrunks == nil {
		return 0, false, nil
	}
	return *value.MaxSIPTrunks, true, nil
}

func (s *Service) RuntimeUsage(ctx context.Context, organizationID uuid.UUID) (RuntimeUsage, error) {
	limits, err := s.Get(ctx, organizationID)
	if err != nil {
		return RuntimeUsage{}, err
	}
	if limits.MonthlyRuntimeMinutes == nil {
		return RuntimeUsage{
			PeriodStart: limits.PeriodStart,
			PeriodEnd:   limits.PeriodEnd,
			Unlimited:   true,
		}, nil
	}
	if s.queries == nil {
		return RuntimeUsage{}, apperror.NewServiceUnavailable("runtime usage store unavailable", nil)
	}
	used, err := s.queries.GetOrganizationRuntimeUsedSeconds(ctx, sqlc.GetOrganizationRuntimeUsedSecondsParams{
		OrganizationID: organizationID,
		PeriodStart:    pgconv.TimeToTimestamptz(limits.PeriodStart),
		PeriodEnd:      pgconv.TimeToTimestamptz(limits.PeriodEnd),
	})
	if err != nil {
		return RuntimeUsage{}, apperror.NewServiceUnavailable("read runtime usage", err)
	}
	return RuntimeUsage{
		PeriodStart:      limits.PeriodStart,
		PeriodEnd:        limits.PeriodEnd,
		UsedSeconds:      used,
		AllowanceSeconds: *limits.MonthlyRuntimeMinutes * 60,
	}, nil
}

// Check only when admitting a NEW session. Do not terminate in-flight calls.
func (s *Service) RequireRuntimeAllowance(ctx context.Context, organizationID uuid.UUID) error {
	usage, err := s.RuntimeUsage(ctx, organizationID)
	if err != nil {
		return err
	}
	if !usage.Unlimited && usage.UsedSeconds >= usage.AllowanceSeconds {
		return apperror.NewPaymentRequired("monthly Voice Agent runtime allowance exhausted")
	}
	return nil
}

func fromPlan(plan plans.Plan) Effective {
	return Effective{
		MaxConcurrentCalls:    positiveLimit(plan.Limits[plans.LimitMaxConcurrentCalls]),
		RetentionDays:         positiveLimit(plan.Limits[plans.LimitRetentionDays]),
		MonthlyRuntimeMinutes: positiveLimit(plan.Limits[plans.LimitMonthlyRuntimeMinutes]),
		MaxVoiceAgents:        positiveLimit(plan.Limits[plans.LimitMaxVoiceAgents]),
		MaxSIPTrunks:          positiveLimit(plan.Limits[plans.LimitMaxSIPTrunks]),
	}
}

func positiveLimit(value int64) *int64 {
	if value < 1 {
		return nil
	}
	return &value
}
