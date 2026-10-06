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
type Message struct{ To, Subject, HTML, Text string }
type Result struct{ MessageID string }

// SendError classifies provider failures without persisting recipient or body data.
type SendError struct {
	Code      string
	Permanent bool
}

func (e *SendError) Error() string { return "email provider: " + e.Code }

type Data struct {
	Code         string    `json:"code,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	Organization string    `json:"organization,omitempty"`
	Inviter      string    `json:"inviter,omitempty"`
	Role         string    `json:"role,omitempty"`
	AcceptURL    string    `json:"accept_url,omitempty"`
}
type Request struct {
	To          string
	Template    string
	Data        Data
	ChallengeID *uuid.UUID
}
type Delivery struct {
	ID                    uuid.UUID
	To, Template, Payload string
	Attempts              int
	ExpiresAt             time.Time
}
