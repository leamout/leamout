package catalog

import (
	"strings"
	"unicode"

	"github.com/google/uuid"
)

func normalizeID(id uuid.UUID) error {
	if id == uuid.Nil {
		return ErrIDRequired
	}
	return nil
}

func normalizeCode(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrCodeRequired
	}
	if strings.IndexFunc(value, unicode.IsSpace) >= 0 {
		return "", ErrInvalidCode
	}
	return value, nil
}
