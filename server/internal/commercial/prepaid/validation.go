package prepaid

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

var operationTypePattern = regexp.MustCompile(`^[a-z0-9]+(?:_[a-z0-9]+)*$`)

func normalizeAuthorization(input AuthorizeInput, now time.Time) (AuthorizeInput, error) {
	input.Currency = strings.ToUpper(strings.TrimSpace(input.Currency))
	input.OperationType = strings.TrimSpace(input.OperationType)
	input.OperationID = strings.TrimSpace(input.OperationID)
	input.ExpiresAt = input.ExpiresAt.UTC()
	if input.OrganizationID == uuid.Nil || len(input.Currency) != 3 || input.AmountMinor <= 0 ||
		!operationTypePattern.MatchString(input.OperationType) || input.OperationID == "" ||
		!input.ExpiresAt.After(now) {
		return AuthorizeInput{}, ErrInvalidManagedOperation
	}
	return input, nil
}

func validateCapture(input CaptureInput) error {
	if input.OrganizationID == uuid.Nil || input.AuthorizationID == uuid.Nil || input.AmountMinor <= 0 {
		return ErrInvalidManagedOperation
	}
	return nil
}

func validateRelease(organizationID, authorizationID uuid.UUID) error {
	if organizationID == uuid.Nil || authorizationID == uuid.Nil {
		return ErrInvalidManagedOperation
	}
	return nil
}
