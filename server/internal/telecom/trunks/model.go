package trunks

import (
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type CreateRequest struct {
	CarrierConnectionID uuid.UUID `json:"carrier_connection_id"`
	Name                string    `json:"name"`
	Direction           *string   `json:"direction,omitempty"`
	Status              *string   `json:"status,omitempty"`
}

type UpdateRequest struct {
	Name      *string `json:"name,omitempty"`
	Direction *string `json:"direction,omitempty"`
	Status    *string `json:"status,omitempty"`
}

type EndpointCreateRequest struct {
	Host      string  `json:"host"`
	Port      *int32  `json:"port,omitempty"`
	Transport *string `json:"transport,omitempty"`
	Direction *string `json:"direction,omitempty"`
	Priority  *int32  `json:"priority,omitempty"`
	Weight    *int32  `json:"weight,omitempty"`
	Enabled   *bool   `json:"enabled,omitempty"`
}

type EndpointUpdateRequest struct {
	Host      *string `json:"host,omitempty"`
	Port      *int32  `json:"port,omitempty"`
	Transport *string `json:"transport,omitempty"`
	Direction *string `json:"direction,omitempty"`
	Priority  *int32  `json:"priority,omitempty"`
	Weight    *int32  `json:"weight,omitempty"`
	Enabled   *bool   `json:"enabled,omitempty"`
}

type Response struct {
	ID                  uuid.UUID `json:"id"`
	OrganizationID      uuid.UUID `json:"organization_id"`
	CarrierConnectionID uuid.UUID `json:"carrier_connection_id"`
	Name                string    `json:"name"`
	Direction           string    `json:"direction"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type EndpointResponse struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	TrunkID        uuid.UUID `json:"trunk_id"`
	Host           string    `json:"host"`
	Port           int32     `json:"port"`
	Transport      string    `json:"transport"`
	Direction      string    `json:"direction"`
	Priority       int32     `json:"priority"`
	Weight         int32     `json:"weight"`
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func response(trunk sqlc.Trunk) Response {
	return Response{trunk.ID, trunk.OrganizationID, trunk.CarrierConnectionID, trunk.Name,
		trunk.Direction, trunk.Status, pgconv.TimestamptzToTime(trunk.CreatedAt), pgconv.TimestamptzToTime(trunk.UpdatedAt)}
}

func endpointResponse(endpoint sqlc.TrunkEndpoint) EndpointResponse {
	return EndpointResponse{endpoint.ID, endpoint.OrganizationID, endpoint.TrunkID, endpoint.Host,
		endpoint.Port, endpoint.Transport, endpoint.Direction, endpoint.Priority, endpoint.Weight,
		endpoint.Enabled, pgconv.TimestamptzToTime(endpoint.CreatedAt), pgconv.TimestamptzToTime(endpoint.UpdatedAt)}
}
