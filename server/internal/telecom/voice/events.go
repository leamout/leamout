package voice

import "time"

type CallEvent struct {
	EventType     CallEventType `json:"event_type"`
	CallID        string        `json:"call_id"`
	ApplicationID string        `json:"application_id"`
	TenantID      string        `json:"tenant_id"`
	From          string        `json:"from"`
	To            string        `json:"to"`
	Direction     CallDirection `json:"direction"`
	Status        CallStatus    `json:"status"`
	DurationSec   int           `json:"duration_seconds,omitempty"`
	OccurredAt    time.Time     `json:"occurred_at"`
	RecordingURL  string        `json:"recording_url,omitempty"`
}

type CallEventType string

const (
	EventCallInitiated CallEventType = "call.initiated"
	EventCallRinging   CallEventType = "call.ringing"
	EventCallAnswered  CallEventType = "call.answered"
	EventCallCompleted CallEventType = "call.completed"
	EventCallFailed    CallEventType = "call.failed"
	EventCallRecording CallEventType = "call.recording.available"
)

type CallDirection string

const (
	DirectionInbound  CallDirection = "inbound"
	DirectionOutbound CallDirection = "outbound"
)

type CallStatus string

const (
	StatusInitiated CallStatus = "initiated"
	StatusRinging   CallStatus = "ringing"
	StatusAnswered  CallStatus = "answered"
	StatusCompleted CallStatus = "completed"
	StatusFailed    CallStatus = "failed"
	StatusBusy      CallStatus = "busy"
	StatusNoAnswer  CallStatus = "no-answer"
)
