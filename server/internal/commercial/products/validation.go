package products

import (
	"errors"
	"strings"
)

var (
	ErrCodeRequired = errors.New("code is required")
	ErrInvalidCode  = errors.New("code must not contain whitespace")
	ErrNameRequired = errors.New("name is required")
)

// Validate checks the invariants shared by product creation and updates.
func Validate(product Product) error {
	if strings.TrimSpace(product.Code) == "" {
		return ErrCodeRequired
	}
	if strings.IndexFunc(product.Code, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }) >= 0 {
		return ErrInvalidCode
	}
	if strings.TrimSpace(product.Name) == "" {
		return ErrNameRequired
	}
	return nil
}
