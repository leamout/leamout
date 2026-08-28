package routing

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrNoRoute          = errors.New("no route available")
	ErrCarrierMismatch  = errors.New("carrier connection does not own phone number")
	ErrTenantMismatch   = errors.New("route organization mismatch")
)

type OutboundRequest struct {
	OrganizationID uuid.UUID
	ApplicationID  *uuid.UUID
	From           string
	To             string
	TrunkID        uuid.UUID
}

type OutboundDecision struct {
	OrganizationID      uuid.UUID
	CarrierConnectionID uuid.UUID
	TrunkID             uuid.UUID
	EndpointID          uuid.UUID
	Host                string
	Port                int32
	Transport           string
	From                string
	To                  string
}

type InboundRequest struct {
	SourceIP     string
	CalledNumber string
	CallerNumber string
}

type InboundDecision struct {
	OrganizationID      uuid.UUID
	CarrierConnectionID uuid.UUID
	PhoneNumberID       uuid.UUID
	VoiceApplicationID  uuid.UUID
	CalledNumber        string
	CallerNumber        string
}
