package subscriptions

import (
	"github.com/google/uuid"
	"github.com/leamout/leamout/server/pkg/apperror"
)

func validateSync(req SyncRequest) error {
	if req.PlanID == uuid.Nil {
		return apperror.NewBadRequest("plan_id is required")
	}
	switch req.Status {
	case StatusTrialing, StatusActive, StatusPastDue, StatusCanceled, StatusIncomplete:
	default:
		return apperror.NewBadRequest("unsupported subscription status")
	}
	if req.PastDueAt != nil || req.GraceEndsAt != nil {
		if req.Status != StatusPastDue || req.PastDueAt == nil || req.GraceEndsAt == nil {
			return apperror.NewBadRequest("payment grace requires past_due status and both timestamps")
		}
		if req.GraceEndsAt.Before(*req.PastDueAt) {
			return apperror.NewBadRequest("grace_ends_at must not be before past_due_at")
		}
	}
	if req.Provider != nil {
		if *req.Provider != ProviderPolar {
			return apperror.NewBadRequest("unsupported subscription provider")
		}
		if req.ProviderCustomerID == nil ||
			req.ProviderSubscriptionID == nil ||
			req.ProviderEventCreatedAt == nil {
			return apperror.NewBadRequest("provider subscription metadata is incomplete")
		}
	}
	if req.CurrentPeriodStart != nil &&
		req.CurrentPeriodEnd != nil &&
		!req.CurrentPeriodEnd.After(*req.CurrentPeriodStart) {
		return apperror.NewBadRequest("current_period_end must be after current_period_start")
	}
	return nil
}
