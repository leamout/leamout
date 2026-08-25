package freeswitch

import (
	"context"
	"fmt"
	"strings"
)

func (c *Client) SetVariable(ctx context.Context, callID, name, value string) error {
	if strings.TrimSpace(callID) == "" || strings.TrimSpace(name) == "" {
		return fmt.Errorf("FreeSWITCH call ID and variable name are required")
	}
	return c.ok(ctx, "uuid_setvar "+callID+" "+name+" "+value)
}

func (c *Client) GetVariable(ctx context.Context, callID, name string) (string, error) {
	if strings.TrimSpace(callID) == "" || strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("FreeSWITCH call ID and variable name are required")
	}

	reply, err := c.Command(ctx, "uuid_getvar "+callID+" "+name)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(reply.Body), nil
}
