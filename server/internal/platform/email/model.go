// Package email owns transactional email rendering and durable delivery.
package email

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Sender interface {
	Send(context.Context, Message) (Result, error)
}
type Message struct {
	To      string
	Subject string
	HTML    string
	Text    string
}
type Result struct {
	MessageID string
}

// SendError classifies provider failures without persisting recipient or body data.
type SendError struct {
	Code      string
	Permanent bool
}

func (e *SendError) Error() string {
	return "email provider: " + e.Code
}

type Data struct {
	UserName      string    `json:"user_name,omitempty"`
	MemberName    string    `json:"member_name,omitempty"`
	EventName     string    `json:"event_name,omitempty"`
	OccurredAt    time.Time `json:"occurred_at,omitempty"`
	KeyName       string    `json:"key_name,omitempty"`
	KeyExpiresAt  time.Time `json:"key_expires_at,omitempty"`
	TrunkName     string    `json:"trunk_name,omitempty"`
	AgentName     string    `json:"agent_name,omitempty"`
	FailureReason string    `json:"failure_reason,omitempty"`
	Code          string    `json:"code,omitempty"`
	ExpiresAt     time.Time `json:"expires_at"`
	Organization  string    `json:"organization,omitempty"`
	Inviter       string    `json:"inviter,omitempty"`
	Role          string    `json:"role,omitempty"`
	AcceptURL     string    `json:"accept_url,omitempty"`
}
type Request struct {
	To              string
	Template        string
	Data            Data
	ExpiresAt       time.Time
	CancellationKey *string
}
type Delivery struct {
	ID        uuid.UUID
	To        string
	Template  string
	Payload   string
	Attempts  int
	ExpiresAt time.Time
}
