package freeswitch

import (
	"context"
	"strings"
)

func (c *Client) Channels(ctx context.Context) ([]Channel, error) {
	reply, err := c.Command(ctx, "show channels as json")
	if err != nil {
		return nil, err
	}
	return decodeRows(reply.Body, func(row map[string]any) Channel {
		return Channel{UUID: stringField(row, "uuid"), Name: stringField(row, "name"), State: stringField(row, "state")}
	})
}

func (c *Client) Calls(ctx context.Context) ([]Call, error) {
	reply, err := c.Command(ctx, "show calls as json")
	if err != nil {
		return nil, err
	}
	return decodeRows(reply.Body, func(row map[string]any) Call {
		return Call{
			UUID:         stringField(row, "uuid"),
			CallerName:   stringField(row, "cid_name"),
			CallerNumber: stringField(row, "cid_num"),
			Destination:  stringField(row, "dest"),
			State:        stringField(row, "state"),
		}
	})
}

func (c *Client) Endpoints(ctx context.Context) ([]Endpoint, error) {
	reply, err := c.Command(ctx, "show endpoints as json")
	if err != nil {
		return nil, err
	}
	return decodeRows(reply.Body, func(row map[string]any) Endpoint {
		return Endpoint{Name: stringField(row, "name"), Type: stringField(row, "type"), Data: stringField(row, "data")}
	})
}

func (c *Client) SofiaStatus(ctx context.Context, profile string) (SIPProfileStatus, error) {
	profile = strings.TrimSpace(profile)
	command := "sofia status"
	if profile != "" {
		command += " profile " + profile
	}

	reply, err := c.Command(ctx, command)
	if err != nil {
		return SIPProfileStatus{}, err
	}
	return SIPProfileStatus{Profile: profile, Raw: reply.Body}, nil
}
