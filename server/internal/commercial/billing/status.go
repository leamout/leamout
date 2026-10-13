package billing

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/commercial/licensing"
	"github.com/leamout/leamout/server/internal/commercial/limits"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
	"github.com/leamout/leamout/server/pkg/apperror"
)

func (s *Service) Status(ctx context.Context, organizationID uuid.UUID) (Status, error) {
	if organizationID == uuid.Nil {
		return Status{}, apperror.NewBadRequest("organization context required")
	}
	result := Status{
		DeploymentMode:    licensing.Cloud,
		BillingConfigured: s.polar.Configured(),
	}
	if s.license.SelfHosted() {
		result.DeploymentMode = licensing.SelfHosted
		result.BillingConfigured = false
		status := s.license.Status()
		result.License = &status
		result.Usage.Unlimited = true
		if status.Claims != nil {
			result.Limits = limits.Effective{
				MaxConcurrentCalls: &status.Claims.MaxConcurrentCalls,
				PeriodStart:        status.Claims.ValidFrom,
				PeriodEnd:          status.Claims.ExpiresAt,
			}
		}
		return result, nil
	}
	sub, err := s.subscriptions.Get(ctx, organizationID)
	if err != nil {
		return Status{}, err
	}
	planCode := "free"
	plan, err := s.plans.GetByCode(ctx, planCode)
	if err != nil {
		return Status{}, err
	}
	if subscriptions.PaidAccessActive(sub, time.Now().UTC()) {
		plan, err = s.plans.GetByID(ctx, sub.PlanID)
		if err != nil {
			return Status{}, err
		}
	}
	result.Plan = &PlanSummary{
		Code:            plan.Code,
		Name:            plan.Name,
		Currency:        plan.Currency,
		AmountMinor:     plan.AmountMinor,
		BillingInterval: plan.BillingInterval,
	}
	if sub != nil {
		result.Subscription = &SubscriptionSummary{
			Status:            sub.Status,
			CurrentPeriodEnd:  sub.CurrentPeriodEnd,
			CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
			PastDueAt:         sub.PastDueAt,
			GraceEndsAt:       sub.GraceEndsAt,
			Provider:          sub.Provider,
		}
		result.PortalAvailable = result.BillingConfigured &&
			sub.Provider != nil &&
			*sub.Provider == providerPolar &&
			sub.ProviderCustomerID != nil
	}
	result.CheckoutAvailable = result.BillingConfigured &&
		(sub == nil ||
			sub.ProviderSubscriptionID == nil ||
			sub.Status == subscriptions.StatusCanceled)
	if s.limits == nil {
		return Status{}, apperror.NewServiceUnavailable("billing limit resolver is unavailable", nil)
	}
	result.Limits, err = s.limits.Get(ctx, organizationID)
	if err != nil {
		return Status{}, err
	}
	result.Usage, err = s.limits.RuntimeUsage(ctx, organizationID)
	return result, err
}
