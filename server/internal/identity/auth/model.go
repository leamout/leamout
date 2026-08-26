package auth

import "github.com/google/uuid"

// -----------------------------------------------------------------------------
// Requests
// -----------------------------------------------------------------------------

type startRequest struct {
	Email string `json:"email"`
}

type passwordLoginRequest struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Password      string    `json:"password"`
}

type sendOTPRequest struct {
	TransactionID uuid.UUID `json:"transaction_id"`
}

type verifyOTPRequest struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Code          string    `json:"code"`
}

type setPasswordRequest struct {
	Password string `json:"password"`
}

// -----------------------------------------------------------------------------
// Responses
// -----------------------------------------------------------------------------

type startResponse struct {
	TransactionID uuid.UUID `json:"transaction_id"`
	Methods       []string  `json:"methods"`
}

type authenticationResponse struct {
	UserID        uuid.UUID `json:"user_id"`
	SessionExpiry string    `json:"session_expires_at"`
}

type sendOTPResponse struct {
	TransactionID uuid.UUID `json:"transaction_id"`
}
