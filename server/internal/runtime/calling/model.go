package calling

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type OriginateRequest struct {
	ChannelID          string
	CallID             uuid.UUID
	Destination        string
	CallerID           string
	TrunkID            uuid.UUID
	Host               string
	PinnedIP           string
	Port               uint16
	Transport          string
	Privacy            bool
	DTMFMode           string
	MediaEncryption    string
	MaxDurationSeconds int32
}

type OriginateResult struct {
	ChannelID string
}

type TransferRequest struct {
	Destination string
	Dialplan    string
	Context     string
}

type RecordRequest struct {
	Path   string
	Action string
}
type OriginateFailureClass string

const (
	OriginateFailureValidation OriginateFailureClass = "validation"
	OriginateFailureTransport  OriginateFailureClass = "transport"
	OriginateFailureTimeout    OriginateFailureClass = "timeout"
	OriginateFailureUncertain  OriginateFailureClass = "uncertain"
	OriginateFailureSIP        OriginateFailureClass = "sip"
	OriginateFailureCapacity   OriginateFailureClass = "capacity"
	OriginateFailureInternal   OriginateFailureClass = "internal"
)

// OriginateError carries the information required to make a failover decision
// without parsing provider or FreeSWITCH error strings.
type OriginateError struct {
	Class     OriginateFailureClass
	SIPStatus int
	Err       error
}

func (e *OriginateError) Error() string {
	if e == nil {
		return "originate failure"
	}
	if e.SIPStatus != 0 {
		return fmt.Sprintf("originate failure (%s, SIP %d): %v", e.Class, e.SIPStatus, e.Err)
	}
	return fmt.Sprintf("originate failure (%s): %v", e.Class, e.Err)
}

func (e *OriginateError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewSIPOriginateError(status int, err error) error {
	if err == nil {
		err = errors.New("SIP originate failed")
	}
	return &OriginateError{Class: OriginateFailureSIP, SIPStatus: status, Err: err}
}

var (
	ErrChannelUnavailable  = errors.New("active call channel unavailable")
	ErrAdmissionCPS        = errors.New("trunk CPS limit exceeded")
	ErrAdmissionConcurrent = errors.New("trunk concurrent call limit exceeded")
	ErrPlanConcurrent      = errors.New("organization concurrent call limit exceeded")
)
var (
	ErrUnsupportedEvent    = errors.New("unsupported FreeSWITCH call event")
	ErrUncorrelatedEvent   = errors.New("uncorrelated FreeSWITCH call event")
	ErrNotInboundAdmission = errors.New("not an inbound FreeSWITCH admission event")
)

type LifecycleType string

const (
	LifecycleInitiated LifecycleType = "initiated"
	LifecycleRinging   LifecycleType = "ringing"
	LifecycleAnswered  LifecycleType = "answered"
	LifecycleActive    LifecycleType = "active"
	LifecycleHeld      LifecycleType = "held"
	LifecycleResumed   LifecycleType = "resumed"
	LifecycleCompleted LifecycleType = "completed"
	LifecycleFailed    LifecycleType = "failed"
	LifecycleCancelled LifecycleType = "cancelled"
)

type LifecycleEvent struct {
	CallID       uuid.UUID
	ChannelID    string
	SIPCallID    string
	Type         LifecycleType
	OccurredAt   time.Time
	HangupReason *string
}

type InboundEvent struct {
	ChannelID           string
	SIPCallID           string
	OrganizationID      uuid.UUID
	VoiceAgentID        uuid.UUID
	PhoneNumberID       uuid.UUID
	VoiceAgentBindingID uuid.UUID
	TrunkID             uuid.UUID
	FromURI             string
	ToURI               string
	OccurredAt          time.Time
}
