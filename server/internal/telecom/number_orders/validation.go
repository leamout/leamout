package number_orders

import (
	"strings"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

func validateOrganizationID(id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.NewBadRequest("organization context required")
	}
	return nil
}

func validateOrderID(id uuid.UUID) error {
	if id == uuid.Nil {
		return apperror.NewBadRequest("number_order_id is required")
	}
	return nil
}

func normalizeSelectionID(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", apperror.NewBadRequest("selection_id is required")
	}
	if !strings.HasPrefix(value, "sel_") || len(value) > 128 {
		return "", apperror.NewBadRequest("invalid selection_id")
	}
	return value, nil
}
