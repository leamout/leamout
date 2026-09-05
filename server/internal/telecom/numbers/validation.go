package numbers

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

var e164 = regexp.MustCompile(`^\+[1-9][0-9]{6,14}$`)
var country = regexp.MustCompile(`^[A-Z]{2}$`)
var numberContains = regexp.MustCompile(`^[0-9]{0,15}$`)

func validateOrganizationID(id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.NewBadRequest("organization_id is required")
	}
	return nil
}

func validateNumberID(id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.NewBadRequest("number id is required")
	}
	return nil
}

func normalizeNumber(v string) (string, error) {
	v = strings.TrimSpace(v)
	if !e164.MatchString(v) {
		return "", apperror.NewBadRequest("number must be in E.164 format")
	}
	return v, nil
}

func normalizeCountryCode(v string) (string, error) {
	v = strings.ToUpper(strings.TrimSpace(v))
	if !country.MatchString(v) {
		return "", apperror.NewBadRequest("country_code must be a two-letter ISO country code")
	}
	return v, nil
}

func normalizeNumberContains(v string) (string, error) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "+")
	if !numberContains.MatchString(v) {
		return "", apperror.NewBadRequest("contains must contain only digits")
	}
	return v, nil
}
