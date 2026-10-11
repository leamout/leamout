package invitations

import (
	"time"

	"github.com/google/uuid"
)

type CreateRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type AcceptRequest struct {
	Token string `json:"token"`
}

type Response struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	ExpiresAt time.Time `json:"expires_at"`
}
