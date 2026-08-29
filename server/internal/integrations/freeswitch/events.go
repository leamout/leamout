package freeswitch

import (
	"context"
	"fmt"
	"net/url"
	"strings"
)

type EventHandler func(context.Context, Event) error

func (c *Client) Subscribe(ctx context.Context, format EventFormat, events []string, handler EventHandler) error {
	if handler == nil {
		return fmt.Errorf("FreeSWITCH event handler is required")
	}
	if format != EventFormatPlain {
		return fmt.Errorf("unsupported FreeSWITCH event format: %q", format)
	}

	clean := make([]string, 0, len(events))
	for _, event := range events {
		event = strings.TrimSpace(event)
		if event != "" {
			clean = append(clean, event)
		}
	}

	command := "event " + string(format)
	if len(clean) > 0 {
		command += " " + strings.Join(clean, " ")
	}

	frame, err := c.command(ctx, command)
	if err != nil {
		return err
	}
	if !frame.OK() {
		return fmt.Errorf("FreeSWITCH event subscription failed: %s", frame.ReplyText())
	}

	c.handlersMu.Lock()
	c.handlers = append(c.handlers, handler)
	c.handlersMu.Unlock()

	return nil
}

func plainEventHeader(body, name string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimRight(line, "\r")
		key, value, ok := strings.Cut(line, ": ")
		if !ok || key != name {
			continue
		}
		decoded, err := url.PathUnescape(value)
		if err == nil {
			return decoded
		}
		return value
	}
	return ""
}
