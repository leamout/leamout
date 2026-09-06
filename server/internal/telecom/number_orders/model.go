package number_orders

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

var (
	ErrSelectionNotFound          = errors.New("number selection not found or expired")
	ErrSelectionUnavailable       = errors.New("number selection is no longer available")
	ErrProviderRoutingUnavailable = errors.New("managed number provider routing is not configured")
)

type CreateRequest struct {
	SelectionID string `json:"selection_id"`
}

type OrderError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

type Response struct {
	ID            uuid.UUID   `json:"id"`
	Number        string      `json:"number"`
	CountryCode   string      `json:"country_code"`
	Status        string      `json:"status"`
	PhoneNumberID *uuid.UUID  `json:"phone_number_id"`
	Error         *OrderError `json:"error"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// ProviderOperationRequest is the internal request used by direct managed-
// provider execution. Provider inventory, product, and routing identifiers stay
// in the durable operation journal rather than the customer number_orders row.
// Transit execution uses its opaque selection contract instead of these
// provider-specific fields.
type ProviderOperationRequest struct {
	SelectionID               string    `json:"selection_id"`
	Provider                  string    `json:"provider"`
	ProviderInventoryID       string    `json:"available_did_id"`
	ProviderProductID         string    `json:"sku_id"`
	Number                    string    `json:"number"`
	CountryCode               string    `json:"country_code"`
	CarrierConnectionID       uuid.UUID `json:"carrier_connection_id"`
	ProviderRoutingResourceID string    `json:"provider_routing_resource_id"`
}

type ProviderOrderRequest struct {
	ProviderInventoryID string
	ProviderProductID   string
	ExternalReferenceID string
}

type ProviderOrder struct {
	ID     string
	Status string
}

type ProviderNumber struct {
	ID                string
	Number            string
	RoutingResourceID string
}

func response(order sqlc.NumberOrder) Response {
	result := Response{
		ID:            order.ID,
		Number:        order.Number,
		CountryCode:   order.CountryCode,
		Status:        order.Status,
		PhoneNumberID: order.PhoneNumberID,
		CreatedAt:     pgconv.TimestamptzToTime(order.CreatedAt),
		UpdatedAt:     pgconv.TimestamptzToTime(order.UpdatedAt),
	}
	if order.ErrorMessage != nil {
		result.Error = &OrderError{Message: *order.ErrorMessage}
		if order.ErrorCode != nil {
			result.Error.Code = *order.ErrorCode
		}
	}
	return result
}
