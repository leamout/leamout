package freeswitch

import (
	"context"
	"fmt"
	"strings"
)

type EventHandler func(context.Context, Event) error

func (c *Client) Subscribe(ctx context.Context, format EventFormat, events []string, handler EventHandler) error {
	if handler == nil {
		return fmt.Errorf("FreeSWITCH event handler is required")
	}
	if format != EventFormatPlain && format != EventFormatJSON {
		return fmt.Errorf("unsupported FreeSWITCH event format: %q", format)
	}

	command := "event " + string(format)
	if len(events) > 0 {
		clean := make([]string, 0, len(events))
		for _, event := range events {
			event = strings.TrimSpace(event)
			if event == "" {
				continue
			}
			clean = append(clean, event)
		}
		if len(clean) > 0 {
			command += " " + strings.Join(clean, " ")
		}
	}

	frame, err := c.command(ctx, command)
	if err != nil {
		return err
	}
	if !frame.OK() {
		return fmt.Errorf("FreeSWITCH event subscription failed: %s", frame.ReplyText())
	}

	return nil
}
