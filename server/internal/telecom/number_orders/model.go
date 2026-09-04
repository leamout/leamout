package number_orders

import (
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type Status string

const (
	StatusPending     Status = "pending"
	StatusPurchasing  Status = "purchasing"
	StatusPurchased   Status = "purchased"
	StatusPersisting  Status = "persisting"
	StatusConfiguring Status = "configuring"
	StatusCompleted   Status = "completed"
	StatusFailed      Status = "failed"
)

type Stage string

const (
	StagePurchasing  Stage = "purchasing"
	StagePersisting  Stage = "persisting"
	StageConfiguring Stage = "configuring"
)

// CreateInput is trusted provisioning input. Provider inventory/product
// identifiers are intentionally not part of the public customer API.
type CreateInput struct {
	OrganizationID      uuid.UUID
	ProviderID          uuid.UUID
	ProviderInventoryID string
	ProviderProductID   string
	Number              string
	CountryCode         string
}

type Failure struct {
	Stage   Stage
	Code    string
	Message string
}

// Response exposes customer-relevant order state without leaking upstream
// provider inventory, SKU, order, or resource identifiers.
type Response struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	Number         string     `json:"number"`
	CountryCode    string     `json:"country_code"`
	Status         Status     `json:"status"`
	FailedStage    *Stage     `json:"failed_stage,omitempty"`
	ErrorCode      *string    `json:"error_code,omitempty"`
	ErrorMessage   *string    `json:"error_message,omitempty"`
	PhoneNumberID  *uuid.UUID `json:"phone_number_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func response(order sqlc.NumberOrder) Response {
	result := Response{
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
	if order.FailedStage != nil {
		stage := Stage(*order.FailedStage)
		result.FailedStage = &stage
	}
	return result
}
