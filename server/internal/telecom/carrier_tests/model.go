package carrier_tests

import (
	"time"

	"github.com/google/uuid"
)

type Request struct {
	TrunkID uuid.UUID `json:"trunk_id"`
	From    string    `json:"from"`
	To      string    `json:"to"`
}

type Result struct {
	ID                  uuid.UUID  `json:"id"`
	OrganizationID      uuid.UUID  `json:"organization_id"`
	CarrierConnectionID uuid.UUID  `json:"carrier_connection_id"`
	TrunkID             uuid.UUID  `json:"trunk_id"`
	TrunkEndpointID     *uuid.UUID `json:"trunk_endpoint_id,omitempty"`
	ActorType           string     `json:"actor_type"`
	ActorID             uuid.UUID  `json:"actor_id"`
	From                string     `json:"from"`
	To                  string     `json:"to"`
	Status              string     `json:"status"`
	SIPCallID           *string    `json:"sip_call_id,omitempty"`
	ResponseCode        *string    `json:"response_code,omitempty"`
	ErrorMessage        *string    `json:"error_message,omitempty"`
	StartedAt           time.Time  `json:"started_at"`
	AnsweredAt          *time.Time `json:"answered_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
}
