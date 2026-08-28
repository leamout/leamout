package numbers

import (
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type CreateRequest struct {
	Number       string `json:"number"`
	CountryCode  string `json:"country_code"`
	VoiceEnabled *bool  `json:"voice_enabled,omitempty"`
	SMSEnabled   *bool  `json:"sms_enabled,omitempty"`
}

type UpdateRequest struct {
	CountryCode  *string `json:"country_code,omitempty"`
	VoiceEnabled *bool   `json:"voice_enabled,omitempty"`
	SMSEnabled   *bool   `json:"sms_enabled,omitempty"`
}

type Response struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	Number         string     `json:"number"`
	CountryCode    string     `json:"country_code"`
	ProviderID     *uuid.UUID `json:"provider_id,omitempty"`
	VoiceEnabled   bool       `json:"voice_enabled"`
	SMSEnabled     bool       `json:"sms_enabled"`
	Status         string     `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

func response(number sqlc.PhoneNumber) Response {
	return Response{
		ID:             number.ID,
		OrganizationID: number.OrganizationID,
		Number:         number.Number,
		CountryCode:    number.CountryCode,
		ProviderID:     number.ProviderID,
		VoiceEnabled:   number.VoiceEnabled,
		SMSEnabled:     number.SmsEnabled,
		Status:         number.Status,
		CreatedAt:      pgconv.TimestamptzToTime(number.CreatedAt),
		UpdatedAt:      pgconv.TimestamptzToTime(number.UpdatedAt),
	}
}
