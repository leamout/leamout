package nats

import (
	"context"
	"fmt"
	"strings"

	natsjs "github.com/nats-io/nats.go/jetstream"
)

func (c *Client) CreateOrUpdateConsumer(ctx context.Context, stream string, config natsjs.ConsumerConfig) (natsjs.Consumer, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}
	if ctx == nil {
		return nil, fmt.Errorf("NATS context is required")
	}
	stream = strings.TrimSpace(stream)
	if stream == "" {
		return nil, fmt.Errorf("NATS stream name is required")
	}
	if strings.TrimSpace(config.Name) == "" && strings.TrimSpace(config.Durable) == "" {
		return nil, fmt.Errorf("NATS consumer name or durable name is required")
	}

	consumer, err := c.jetStream.CreateOrUpdateConsumer(ctx, stream, config)
	if err != nil {
		return nil, fmt.Errorf("create or update NATS consumer on %s: %w", stream, err)
	}

	return consumer, nil
}
