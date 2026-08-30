package plans

import (
	"errors"
	"strings"
)

var (
	ErrProductRequired = errors.New("product_id is required")
	ErrCodeRequired    = errors.New("code is required")
	ErrInvalidCode     = errors.New("code must not contain whitespace")
	ErrNameRequired    = errors.New("name is required")
)

// Validate checks the invariants shared by plan creation and updates.
func Validate(plan Plan) error {
	if strings.TrimSpace(plan.ProductID) == "" {
		return ErrProductRequired
	}
	if strings.TrimSpace(plan.Code) == "" {
		return ErrCodeRequired
	}
	if strings.IndexFunc(plan.Code, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }) >= 0 {
		return ErrInvalidCode
	}
	if strings.TrimSpace(plan.Name) == "" {
		return ErrNameRequired
	}
	return nil
}
