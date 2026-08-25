package nats

import (
	"context"
	"fmt"
	"strings"

	natsgo "github.com/nats-io/nats.go"
	natsjs "github.com/nats-io/nats.go/jetstream"
)

const MaxMessageSize = 64 * 1024

func (c *Client) Publish(ctx context.Context, subject string, payload []byte) error {
	return c.PublishWithOptions(ctx, subject, payload, nil, "")
}

func (c *Client) PublishWithOptions(ctx context.Context, subject string, payload []byte, headers map[string]string, messageID string) error {
	if err := c.validate(); err != nil {
		return err
	}
	if ctx == nil {
		return fmt.Errorf("NATS context is required")
	}
	subject = strings.TrimSpace(subject)
	if subject == "" {
		return fmt.Errorf("NATS subject is required")
	}
	if len(payload) > MaxMessageSize {
		return fmt.Errorf("NATS message payload is %d bytes; maximum is %d", len(payload), MaxMessageSize)
	}

	msg := &natsgo.Msg{Subject: subject, Data: payload}
	for key, value := range headers {
		key = strings.TrimSpace(key)
		if key != "" {
			msg.Header.Set(key, value)
		}
	}

	var options []natsjs.PublishOpt
	if messageID = strings.TrimSpace(messageID); messageID != "" {
		options = append(options, natsjs.WithMsgID(messageID))
	}

	if _, err := c.jetStream.PublishMsg(ctx, msg, options...); err != nil {
		return fmt.Errorf("publish NATS message to %s: %w", subject, err)
	}

	return nil
}
