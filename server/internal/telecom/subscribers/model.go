package subscribers

import (
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
	"time"
)

type CreateRequest struct {
	SIPDomainID uuid.UUID `json:"sip_domain_id"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	DisplayName *string   `json:"display_name,omitempty"`
}
type UpdateRequest struct {
	DisplayName *string `json:"display_name,omitempty"`
}
type RotateCredentialsRequest struct {
	Password string `json:"password"`
}
type Response struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	SIPDomainID    uuid.UUID `json:"sip_domain_id"`
	Username       string    `json:"username"`
	Domain         string    `json:"domain"`
	DisplayName    *string   `json:"display_name,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func response(v sqlc.Subscriber) Response {
	return Response{v.ID, v.OrganizationID, v.SipDomainID, v.Username, v.Domain, v.DisplayName, v.Status, pgconv.TimestamptzToTime(v.CreatedAt), pgconv.TimestamptzToTime(v.UpdatedAt)}
}
