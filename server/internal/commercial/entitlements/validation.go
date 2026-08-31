package entitlements

import (
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

func validateID(id uuid.UUID, required error) error {
	if id == uuid.Nil {
		return required
	}
	return nil
}

func normalizeKey(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrKeyRequired
	}
	if strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", ErrInvalidKey
	}
	return value, nil
}

func normalizeKind(kind Kind) (Kind, error) {
	switch kind {
	case KindFeature, KindLimit:
		return kind, nil
	default:
		return "", ErrInvalidKind
	}
}

func normalizeCreate(input CreateInput) (CreateInput, error) {
	key, err := normalizeKey(input.Key)
	if err != nil {
		return CreateInput{}, err
	}
	kind, err := normalizeKind(input.Kind)
	if err != nil {
		return CreateInput{}, err
	}
	input.Key = key
	input.Kind = kind

	switch kind {
	case KindFeature:
		if input.Enabled == nil || input.LimitValue != nil {
			return CreateInput{}, ErrFeatureValueRequired
		}
	case KindLimit:
		if input.LimitValue == nil || input.Enabled != nil {
			return CreateInput{}, ErrLimitValueRequired
		}
		if *input.LimitValue < 0 {
			return CreateInput{}, ErrInvalidLimit
		}
	}

	if err := validatePeriod(input.StartsAt, input.ExpiresAt); err != nil {
		return CreateInput{}, err
	}
	return input, nil
}

func validatePeriod(startsAt, expiresAt *time.Time) error {
	if startsAt != nil && expiresAt != nil && expiresAt.Before(*startsAt) {
		return ErrInvalidPeriod
	}
	return nil
}

func effectiveAt(entitlement Entitlement, at time.Time) bool {
	if entitlement.StartsAt != nil && at.Before(*entitlement.StartsAt) {
		return false
	}
	if entitlement.ExpiresAt != nil && at.After(*entitlement.ExpiresAt) {
		return false
	}
	return true
}
