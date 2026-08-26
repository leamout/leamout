package auth

import (
	"strings"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func normalizeOTP(code string) string {
	return strings.TrimSpace(code)
}

func validateStartRequest(req startRequest) error {
	if normalizeEmail(req.Email) == "" {
		return apperror.NewBadRequest("email is required")
	}

	return nil
}

func validatePasswordLoginRequest(
	req passwordLoginRequest,
) error {
	if req.TransactionID == uuid.Nil {
		return apperror.NewBadRequest(
			"invalid transaction_id",
		)
	}

	if req.Password == "" {
		return apperror.NewBadRequest(
			"password is required",
		)
	}

	return nil
}
