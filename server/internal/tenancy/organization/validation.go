package organization

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

func validateOrganizationID(id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.NewBadRequest("organization_id is required")
	}

	return nil
}

func normalizeRequiredString(value string, field string) (string, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return "", apperror.NewBadRequest(field + " is required")
	}

	return normalized, nil
}

func normalizeSlug(value string) (string, error) {
	slug := strings.ToLower(strings.TrimSpace(value))
	if !slugPattern.MatchString(slug) {
		return "", apperror.NewBadRequest("slug must be 3-63 characters and contain only lowercase letters, numbers, and hyphens")
	}

	return slug, nil
}
