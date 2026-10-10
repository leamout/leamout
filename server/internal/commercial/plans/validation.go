package plans

import (
	"strings"

	"github.com/leamout/leamout/server/internal/commercial/entitlements"
	"github.com/leamout/leamout/server/pkg/apperror"
)

func validateUpsert(req UpsertRequest) (UpsertRequest, error) {
	req.Code = strings.TrimSpace(req.Code)
	req.Name = strings.TrimSpace(req.Name)
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	req.BillingInterval = strings.TrimSpace(req.BillingInterval)
	req.Status = strings.TrimSpace(req.Status)

	if req.Code == "" {
		return UpsertRequest{}, apperror.NewBadRequest("plan code is required")
	}
	if req.Name == "" {
		return UpsertRequest{}, apperror.NewBadRequest("plan name is required")
	}
	if len(req.Currency) != 3 {
		return UpsertRequest{}, apperror.NewBadRequest("plan currency must be a 3-letter code")
	}
	if req.AmountMinor < 0 {
		return UpsertRequest{}, apperror.NewBadRequest("plan amount must be non-negative")
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
