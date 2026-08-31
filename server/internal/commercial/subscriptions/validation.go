package subscriptions

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

func validateID(id uuid.UUID, missing error) error {
	if id == uuid.Nil {
		return missing
	}
	return nil
}

func normalizeStatus(status Status) (Status, error) {
	value := Status(strings.ToLower(strings.TrimSpace(string(status))))
	switch value {
	case StatusPending, StatusActive, StatusPastDue, StatusCancelled, StatusExpired:
		return value, nil
	default:
		return "", ErrInvalidStatus
	}
}

func normalizeInitialStatus(status *Status) (*Status, error) {
	if status == nil {
		return nil, nil
	}
	value, err := normalizeStatus(*status)
	if err != nil {
		return nil, err
	}
	if value != StatusPending && value != StatusActive {
		return nil, ErrInvalidInitialStatus
	}
	return &value, nil
}

func validateTransition(from, to Status) error {
	if from == to {
		return nil
	}
	switch from {
	case StatusPending:
		if to == StatusActive {
			return nil
		}
	case StatusActive:
		if to == StatusPastDue || to == StatusCancelled || to == StatusExpired {
			return nil
		}
	case StatusPastDue:
		if to == StatusActive || to == StatusCancelled || to == StatusExpired {
			return nil
		}
	case StatusCancelled, StatusExpired:
		return ErrInvalidTransition
	}
	return ErrInvalidTransition
}

func normalizeProvider(reference ProviderReference) (ProviderReference, error) {
	provider := strings.ToLower(strings.TrimSpace(reference.Provider))
	if provider == "" {
		return ProviderReference{}, ErrProviderRequired
	}
	if strings.IndexFunc(provider, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }) >= 0 {
		return ProviderReference{}, ErrInvalidProvider
	}
	providerID := strings.TrimSpace(reference.SubscriptionID)
	if providerID == "" {
		return ProviderReference{}, ErrProviderIDRequired
	}
	return ProviderReference{Provider: provider, SubscriptionID: providerID}, nil
}

func validatePeriod(startsAt time.Time, renewsAt, endsAt *time.Time) error {
	if startsAt.IsZero() {
		return ErrInvalidPeriod
	}
	if renewsAt != nil && renewsAt.Before(startsAt) {
		return ErrInvalidPeriod
	}
	if endsAt != nil && endsAt.Before(startsAt) {
		return ErrInvalidPeriod
	}
	if renewsAt != nil && endsAt != nil && renewsAt.After(*endsAt) {
		return ErrInvalidPeriod
	}
	return nil
}

func normalizeCreate(input CreateInput, now time.Time) (CreateInput, error) {
	if err := validateID(input.PlanID, ErrPlanIDRequired); err != nil {
		return CreateInput{}, err
	}
	status, err := normalizeInitialStatus(input.Status)
	if err != nil {
		return CreateInput{}, err
	}
	input.Status = status
	if input.StartsAt == nil {
		value := now.UTC()
		input.StartsAt = &value
	} else {
		value := input.StartsAt.UTC()
		input.StartsAt = &value
	}
	if input.RenewsAt != nil {
		value := input.RenewsAt.UTC()
		input.RenewsAt = &value
	}
	if input.EndsAt != nil {
		value := input.EndsAt.UTC()
		input.EndsAt = &value
	}
	if err := validatePeriod(*input.StartsAt, input.RenewsAt, input.EndsAt); err != nil {
		return CreateInput{}, err
	}
	if input.Provider != nil {
		reference, err := normalizeProvider(*input.Provider)
		if err != nil {
			return CreateInput{}, err
		}
		input.Provider = &reference
	}
	return input, nil
}

func normalizePeriodUpdate(current Subscription, input PeriodUpdate) (PeriodUpdate, error) {
	if input.RenewsAt == nil && input.EndsAt == nil {
		return PeriodUpdate{}, ErrNoChanges
	}
	renewsAt := current.RenewsAt
	endsAt := current.EndsAt
	if input.RenewsAt != nil {
		value := input.RenewsAt.UTC()
		input.RenewsAt = &value
		renewsAt = &value
	}
	if input.EndsAt != nil {
		value := input.EndsAt.UTC()
		input.EndsAt = &value
		endsAt = &value
	}
	if err := validatePeriod(current.StartsAt, renewsAt, endsAt); err != nil {
		return PeriodUpdate{}, err
	}
	return input, nil
}
