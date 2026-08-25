package opensips

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const DefaultEventPath = "/events"

// Events consumes newline-delimited JSON events from an OpenSIPS event endpoint.
// The integration deliberately exposes the decoded event and leaves domain
// interpretation to the caller.
func (c *Client) Events(ctx context.Context, path string, handler func(context.Context, Event) error) error {
	if err := c.validate(ctx); err != nil {
		return err
	}
	if handler == nil {
		return fmt.Errorf("OpenSIPS event handler is required")
	}
	if strings.TrimSpace(path) == "" {
		path = DefaultEventPath
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/"+strings.TrimLeft(path, "/"), nil)
	if err != nil {
		return fmt.Errorf("create OpenSIPS event request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect to OpenSIPS events: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("OpenSIPS events returned HTTP %d", resp.StatusCode)
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var event Event
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("decode OpenSIPS event: %w", err)
		}
		if err := handler(ctx, event); err != nil {
			return fmt.Errorf("handle OpenSIPS event %q: %w", event.Name, err)
		}
	}
}
