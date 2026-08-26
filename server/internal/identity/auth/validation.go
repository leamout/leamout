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

func normalizeTransactionID(value string) string {
	return strings.TrimSpace(value)
}

func normalizePassword(password string) string {
	return strings.TrimSpace(password)
}

func parseUUID(value string) (uuid.UUID, error) {
	return uuid.Parse(normalizeTransactionID(value))
}

func validateStartRequest(req startRequest) error {
	if normalizeEmail(req.Email) == "" {
		return apperror.NewBadRequest("email is required")
	}

	return nil
}

func validatePasswordLoginRequest(req passwordLoginRequest) error {
	if normalizeTransactionID(req.TransactionID) == "" {
		return apperror.NewBadRequest("transaction_id is required")
	}

	if _, err := parseUUID(req.TransactionID); err != nil {
		return apperror.NewBadRequest("invalid transaction_id")
	}

	if normalizePassword(req.Password) == "" {
		return apperror.NewBadRequest("password is required")
	}

	return nil
}

func validateSendOTPRequest(req sendOTPRequest) error {
	if normalizeTransactionID(req.TransactionID) == "" {
		return apperror.NewBadRequest("transaction_id is required")
	}

	if _, err := parseUUID(req.TransactionID); err != nil {
		return apperror.NewBadRequest("invalid transaction_id")
	}

	return nil
}

func validateVerifyOTPRequest(req verifyOTPRequest) error {
	if normalizeTransactionID(req.TransactionID) == "" {
		return apperror.NewBadRequest("transaction_id is required")
	}

	if _, err := parseUUID(req.TransactionID); err != nil {
		return apperror.NewBadRequest("invalid transaction_id")
	}

	if normalizeOTP(req.Code) == "" {
		return apperror.NewBadRequest("code is required")
	}

	return nil
}

func validateSetPasswordRequest(req setPasswordRequest) error {
	if normalizePassword(req.Password) == "" {
		return apperror.NewBadRequest("password is required")
	}

	return nil
}
