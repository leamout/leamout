package number_orders

import (
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

// CreateInput is trusted provisioning input. Provider inventory/product
// identifiers are persisted as immutable selection context but intentionally
// omitted from the public customer response.
type CreateInput struct {
	OrganizationID      uuid.UUID
	ProviderID          uuid.UUID
	ProviderInventoryID string
	ProviderProductID   string
	Number              string
	CountryCode         string
}

// Response exposes only customer-relevant order state.
type Response struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	Number         string     `json:"number"`
	CountryCode    string     `json:"country_code"`
	Status         Status     `json:"status"`
	ErrorCode      *string    `json:"error_code,omitempty"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
	PhoneNumberID  *uuid.UUID `json:"phone_number_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func response(order sqlc.NumberOrder) Response {
	return Response{
		ID:             order.ID,
		OrganizationID: order.OrganizationID,
		Number:         order.Number,
		CountryCode:    order.CountryCode,
		Status:         Status(order.Status),
		ErrorCode:      order.ErrorCode,
		ErrorMessage:   order.ErrorMessage,
		PhoneNumberID:  order.PhoneNumberID,
		CreatedAt:      pgconv.TimestamptzToTime(order.CreatedAt),
		UpdatedAt:      pgconv.TimestamptzToTime(order.UpdatedAt),
	}
}
