package notifications

import "context"

// Email is a fully rendered transactional email ready for delivery.
type Email struct {
	To      string
	From    string
	ReplyTo string
	Subject string
	HTML    string
	Text    string
}

// Delivery identifies a message accepted by an email provider.
type Delivery struct {
	Provider  string
	MessageID string
}

// Sender transports rendered transactional email.
type Sender interface {
	Send(ctx context.Context, email Email) (Delivery, error)
}
