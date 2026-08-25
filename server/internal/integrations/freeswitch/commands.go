package freeswitch

import (
	"context"
	"encoding/json"
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

func (c *Client) ok(ctx context.Context, command string) error {
	reply, err := c.Command(ctx, command)
	if err != nil {
		return err
	}
	body := strings.TrimSpace(reply.Body)
	if strings.HasPrefix(body, "-ERR") {
		return fmt.Errorf("FreeSWITCH command failed: %s", body)
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

func variables(values map[string]string) string {
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
		parts = append(parts, key+"="+values[key])
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func decodeRows[T any](body string, mapRow func(map[string]any) T) ([]T, error) {
	rows, err := responseRows(body)
	if err != nil {
		return nil, err
	}

	result := make([]T, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapRow(row))
	}
	return result, nil
}

func responseRows(body string) ([]map[string]any, error) {
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(strings.TrimSpace(body)), &raw); err != nil {
		return nil, fmt.Errorf("decode FreeSWITCH JSON response: %w", err)
	}

	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err == nil {
		return rows, nil
	}

	var envelope struct {
		Rows []map[string]any `json:"rows"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, fmt.Errorf("decode FreeSWITCH JSON response envelope: %w", err)
	}
	if envelope.Rows == nil {
		return nil, fmt.Errorf("decode FreeSWITCH JSON response: missing rows field")
	}
	return envelope.Rows, nil
}

func stringField(row map[string]any, key string) string {
	value, ok := row[key]
	if !ok {
		return ""
	}
	return fmt.Sprint(value)
}
