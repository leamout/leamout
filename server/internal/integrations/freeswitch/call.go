package freeswitch

import (
	"context"
	"fmt"
	"strings"
)

func (c *Client) Originate(ctx context.Context, req OriginateRequest) (Call, error) {
	if strings.TrimSpace(req.Endpoint) == "" || strings.TrimSpace(req.Destination) == "" {
		return Call{}, fmt.Errorf("FreeSWITCH originate endpoint and destination are required")
	}

	command := "originate"
	if prefix := formatVariables(req.Variables); prefix != "" {
		command += " " + prefix
	}
	if req.CallerID != "" {
		callerIDVar := "origination_caller_id_number=" + req.CallerID
		if strings.HasSuffix(command, "}") {
			command = strings.TrimSuffix(command, "}") + "," + callerIDVar + "}"
		} else {
			command += " {" + callerIDVar + "}"
		}
	}
	command += " " + req.Endpoint + " " + req.Destination

	reply, err := c.Command(ctx, command)
	if err != nil {
		return Call{}, err
	}
	body := strings.TrimSpace(reply.Body)
	if !strings.HasPrefix(body, "+OK") {
		return Call{}, fmt.Errorf("FreeSWITCH originate failed: %s", body)
	}
	return Call{UUID: strings.TrimSpace(strings.TrimPrefix(body, "+OK")), Destination: req.Destination}, nil
}

func (c *Client) Hangup(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.commandOK(ctx, "uuid_kill "+callID)
}

func (c *Client) Hold(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.commandOK(ctx, "uuid_hold "+callID)
}

func (c *Client) Unhold(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.commandOK(ctx, "uuid_unhold "+callID)
}

func (c *Client) Break(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.commandOK(ctx, "uuid_break "+callID+" all")
}

func (c *Client) Transfer(ctx context.Context, req TransferRequest) error {
	if strings.TrimSpace(req.CallID) == "" || strings.TrimSpace(req.Destination) == "" {
		return fmt.Errorf("FreeSWITCH transfer call ID and destination are required")
	}
	args := []string{req.CallID, req.Destination}
	if req.Dialplan != "" {
		args = append(args, req.Dialplan)
	}
	if req.Context != "" {
		args = append(args, req.Context)
	}
	return c.commandOK(ctx, "uuid_transfer "+strings.Join(args, " "))
}

func (c *Client) Record(ctx context.Context, req RecordRequest) error {
	if strings.TrimSpace(req.CallID) == "" || strings.TrimSpace(req.Path) == "" {
		return fmt.Errorf("FreeSWITCH record call ID and path are required")
	}
	action := strings.TrimSpace(req.Action)
	if action == "" {
		action = "start"
	}
	if action != "start" && action != "stop" {
		return fmt.Errorf("FreeSWITCH record action must be start or stop, got %q", action)
	}
	return c.commandOK(ctx, "uuid_record "+req.CallID+" "+action+" "+req.Path)
}
