package plans

import (
	"strings"

	"github.com/leamout/leamout/server/internal/commercial/entitlements"
	"github.com/leamout/leamout/server/pkg/apperror"
)

func validateUpsert(req UpsertRequest) (UpsertRequest, error) {
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	req.PricingType = strings.TrimSpace(req.PricingType)
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	req.BillingInterval = strings.TrimSpace(req.BillingInterval)
	req.Status = strings.TrimSpace(req.Status)

	if req.Code == "" {
		return UpsertRequest{}, apperror.NewBadRequest("plan code is required")
	}
	if req.Name == "" {
		return UpsertRequest{}, apperror.NewBadRequest("plan name is required")
	}
	switch req.PricingType {
	case PricingTypeFixed:
		if req.AmountMinor == nil {
			return UpsertRequest{}, apperror.NewBadRequest("fixed plan amount is required")
		}
		if *req.AmountMinor < 0 {
			return UpsertRequest{}, apperror.NewBadRequest("plan amount must be non-negative")
		}
	case PricingTypeCustom:
		if req.AmountMinor != nil {
			return UpsertRequest{}, apperror.NewBadRequest("custom plan amount must be omitted")
		}
	default:
		return UpsertRequest{}, apperror.NewBadRequest("unsupported pricing type")
	}
	if len(req.Currency) != 3 {
		return UpsertRequest{}, apperror.NewBadRequest("plan currency must be a 3-letter code")
	}
	if req.BillingInterval != IntervalMonth && req.BillingInterval != IntervalYear {
		return UpsertRequest{}, apperror.NewBadRequest("unsupported billing interval")
	}
	if req.Status != StatusActive && req.Status != StatusInactive {
		return UpsertRequest{}, apperror.NewBadRequest("unsupported plan status")
	}

	if req.Entitlements == nil {
		req.Entitlements = map[string]bool{}
	}
	for capability := range req.Entitlements {
		if !entitlements.Capability(capability).IsValid() {
			return UpsertRequest{}, apperror.NewBadRequest("unsupported plan entitlement")
		}
	}

	if req.Limits == nil {
		req.Limits = map[string]int64{}
	}
	for name, value := range req.Limits {
		switch name {
		case LimitMaxConcurrentCalls, LimitRetentionDays:
			if value <= 0 {
				return UpsertRequest{}, apperror.NewBadRequest("plan limit must be positive")
			}
		default:
			return UpsertRequest{}, apperror.NewBadRequest("unsupported plan limit")
		}
	}

	if req.Description != nil {
		value := strings.TrimSpace(*req.Description)
		if value == "" {
			req.Description = nil
		} else {
			req.Description = &value
		}
	}
	return req, nil
}
