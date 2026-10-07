package invitations

import (
	"github.com/coffeyvidzro/monogo/internal/database/pgconv"
	"github.com/coffeyvidzro/monogo/internal/database/sqlc"
	"github.com/google/uuid"
	"time"
)

type CreateRequest struct {
	Email string `json:"email"`
	Role  string `json:"role"`
}

type AcceptRequest struct {
	Token string `json:"token"`
}

type Response struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Email          string    `json:"email"`
	Role           string    `json:"role"`
	Status         string    `json:"status"`
	ExpiresAt      time.Time `json:"expires_at"`
}

func toResponse(i sqlc.OrganizationInvitation) Response {
	return Response{
		ID: i.ID, OrganizationID: i.OrganizationID, Email: i.Email,
		Role: i.Role, Status: i.Status, ExpiresAt: pgconv.TimestamptzToTime(i.ExpiresAt),
	}
}
