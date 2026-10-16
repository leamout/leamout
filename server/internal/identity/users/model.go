package users

import (
	"time"

	"github.com/google/uuid"
)

// Response is the public representation of an authenticated user.
// Identity provider tokens are never exposed.
type Response struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Name          *string   `json:"name,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
