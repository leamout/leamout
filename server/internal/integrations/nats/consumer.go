package nats

import (
	"context"
	"fmt"
	"strings"

	natsjs "github.com/nats-io/nats.go/jetstream"
)

func (c *Client) CreateOrUpdateConsumer(ctx context.Context, stream string, config natsjs.ConsumerConfig) (natsjs.Consumer, error) {
	if err := c.validateConsumer(ctx, stream, config); err != nil {
		return nil, err
	}

	consumer, err := c.jetStream.CreateOrUpdateConsumer(ctx, stream, config)
	if err != nil {
		return nil, fmt.Errorf("create or update NATS consumer on %q: %w", stream, err)
	}

	return consumer, nil
}

// Consume starts a pull consumer handler and blocks until ctx is cancelled.
// The handler owns acknowledgement and should Ack, Nak, Term, or InProgress each message.
func (c *Client) Consume(ctx context.Context, consumer natsjs.Consumer, handler func(context.Context, natsjs.Msg) error) error {
	if err := c.validate(); err != nil {
		return err
	}
	if ctx == nil {
		return fmt.Errorf("NATS context is required")
	}
	if consumer == nil {
		return fmt.Errorf("NATS consumer is required")
	}
	if handler == nil {
		return fmt.Errorf("NATS consumer handler is required")
	}

	cons, err := consumer.Consume(func(msg natsjs.Msg) {
		if err := handler(ctx, msg); err != nil {
			_ = msg.Nak()
		}
	})
	if err != nil {
		return fmt.Errorf("start NATS consumer: %w", err)
	}
	defer cons.Stop()

	<-ctx.Done()
	return ctx.Err()
}

func (c *Client) validateConsumer(ctx context.Context, stream string, config natsjs.ConsumerConfig) error {
	if err := c.validate(); err != nil {
		return err
	}
	if ctx == nil {
		return fmt.Errorf("NATS context is required")
	}
	if stream = strings.TrimSpace(stream); stream == "" {
		return fmt.Errorf("NATS stream name is required")
	}
	if strings.TrimSpace(config.Name) == "" && strings.TrimSpace(config.Durable) == "" {
		return fmt.Errorf("NATS consumer name or durable name is required")
	}

	return nil
}
