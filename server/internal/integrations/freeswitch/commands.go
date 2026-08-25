package freeswitch

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// Command sends a synchronous API command to FreeSWITCH and returns the response.
func (c *Client) Command(ctx context.Context, command string) (Reply, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return Reply{}, fmt.Errorf("FreeSWITCH command is required")
	}

	frame, err := c.command(ctx, "api "+command)
	if err != nil {
		return Reply{}, err
	}
	if frame.ContentType != ContentTypeAPIResponse {
		return Reply{}, fmt.Errorf("unexpected FreeSWITCH command response: %q", frame.ContentType)
	}

	return Reply{Text: frame.ReplyText(), Body: frame.Body}, nil
}

// BGAPI sends a background API command to FreeSWITCH and returns the job UUID.
func (c *Client) BGAPI(ctx context.Context, command string) (Job, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return Job{}, fmt.Errorf("FreeSWITCH background command is required")
	}

	frame, err := c.command(ctx, "bgapi "+command)
	if err != nil {
		return Job{}, err
	}
	if !frame.OK() {
		return Job{}, fmt.Errorf("FreeSWITCH background command failed: %s", frame.ReplyText())
	}

	return Job{ID: frame.Header("Job-UUID")}, nil
}

func (c *Client) commandOK(ctx context.Context, command string) error {
	reply, err := c.Command(ctx, command)
	if err != nil {
		return err
	}
	if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(reply.Body)), "-ERR") {
		return fmt.Errorf("FreeSWITCH command failed: %s", strings.TrimSpace(reply.Body))
	}
	return nil
}

func requiredArgument(name, value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("FreeSWITCH %s is required", name)
	}
	return value, nil
}

func commandWord(value string) string {
	if strings.ContainsAny(value, " \t\r\n\"'") {
		return `"` + strings.ReplaceAll(value, `"`, `\"`) + `"`
	}
	return value
}

func commandWords(values ...string) string {
	words := make([]string, 0, len(values))
	for _, value := range values {
		words = append(words, commandWord(value))
	}
	return strings.Join(words, " ")
}

func formatVariables(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}

	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+commandWord(values[key]))
	}
	return "{" + strings.Join(parts, ",") + "}"
}
