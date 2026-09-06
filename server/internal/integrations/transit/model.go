package transit

import "github.com/google/uuid"

type AvailableNumberSearchRequest struct {
	CountryCode string `json:"country_code"`
	Contains    string `json:"contains,omitempty"`
}

type AvailableNumber struct {
	SelectionID  string `json:"selection_id"`
	Number       string `json:"number"`
	CountryCode  string `json:"country_code"`
	VoiceEnabled bool   `json:"voice_enabled"`
}

type AvailableNumberSearchResponse struct {
	Numbers []AvailableNumber `json:"numbers"`
}

type ExecuteNumberOrderRequest struct {
	OperationID uuid.UUID `json:"operation_id"`
	SelectionID string    `json:"selection_id"`
}

type ExecuteNumberOrderResponse struct {
	State             string `json:"state"`
	ManagedResourceID string `json:"managed_resource_id,omitempty"`
	Number            string `json:"number,omitempty"`
	CountryCode       string `json:"country_code,omitempty"`
	ErrorCode         string `json:"error_code,omitempty"`
	ErrorMessage      string `json:"error_message,omitempty"`
}
