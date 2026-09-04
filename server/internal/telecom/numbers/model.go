package numbers

import (
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/pgconv"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type ProvisioningMode string

const (
	ProvisioningModeBYOC    ProvisioningMode = "byoc"
	ProvisioningModeManaged ProvisioningMode = "managed"
)

type CreateRequest struct {
	Number       string `json:"number"`
	CountryCode  string `json:"country_code"`
	VoiceEnabled *bool  `json:"voice_enabled,omitempty"`
	SMSEnabled   *bool  `json:"sms_enabled,omitempty"`
}

// ManagedCreateRequest is an internal provisioning input. Provider metadata is
// supplied by trusted provisioning orchestration after the upstream resource
// has been created; it is never accepted from the public numbers API.
type ManagedCreateRequest struct {
	Number              string
	CountryCode         string
	ProviderID          uuid.UUID
	ProviderResourceID  string
	CarrierConnectionID *uuid.UUID
	VoiceEnabled        *bool
	SMSEnabled          *bool
}

type UpdateRequest struct {
	CountryCode  *string `json:"country_code,omitempty"`
	VoiceEnabled *bool   `json:"voice_enabled,omitempty"`
	SMSEnabled   *bool   `json:"sms_enabled,omitempty"`
}

type CarrierConnectionRequest struct {
	CarrierConnectionID uuid.UUID `json:"carrier_connection_id"`
}

type Response struct {
	ID                  uuid.UUID  `json:"id"`
	OrganizationID      uuid.UUID  `json:"organization_id"`
	Number              string     `json:"number"`
	CountryCode         string     `json:"country_code"`
	CarrierConnectionID *uuid.UUID `json:"carrier_connection_id,omitempty"`
	VoiceEnabled        bool       `json:"voice_enabled"`
	SMSEnabled          bool       `json:"sms_enabled"`
	Status              string     `json:"status"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func response(number sqlc.PhoneNumber) Response {
	return Response{
		ID:                  number.ID,
		OrganizationID:      number.OrganizationID,
		Number:              number.Number,
		CountryCode:         number.CountryCode,
		CarrierConnectionID: number.CarrierConnectionID,
		VoiceEnabled:        number.VoiceEnabled,
		SMSEnabled:          number.SmsEnabled,
		Status:              number.Status,
		CreatedAt:           pgconv.TimestamptzToTime(number.CreatedAt),
		UpdatedAt:           pgconv.TimestamptzToTime(number.UpdatedAt),
	}
}
