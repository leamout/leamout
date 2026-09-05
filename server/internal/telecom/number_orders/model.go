package number_orders

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

var (
	ErrSelectionNotFound   = errors.New("number selection not found or expired")
	ErrSelectionUnavailable = errors.New("number selection is no longer available")
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
