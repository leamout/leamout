package freeswitch

import (
	"context"
	"fmt"
	"strings"
)

func (c *Client) Originate(ctx context.Context, req OriginateRequest) (Call, error) {
	if err := req.Validate(); err != nil {
		return Call{}, err
	}

	command := "originate"
	if prefix := formatVariables(req.Variables); prefix != "" {
		command += " " + prefix
	}
	if req.CallerID != "" {
		callerIDVar := "origination_caller_id_number=" + commandWord(req.CallerID)
		if strings.HasSuffix(command, "}") {
			command = strings.TrimSuffix(command, "}") + "," + callerIDVar + "}"
		} else {
			command += " {" + callerIDVar + "}"
		}
	}
	command += " " + commandWords(req.Endpoint, req.Destination)

	reply, err := c.Command(ctx, command)
	if err != nil {
		return Call{}, err
	}
	body := strings.TrimSpace(reply.Body)
	if !strings.HasPrefix(body, "+OK") {
		return Call{}, fmt.Errorf("FreeSWITCH originate failed: %s", body)
	}

	return Call{
		UUID:        strings.TrimSpace(strings.TrimPrefix(body, "+OK")),
		Destination: req.Destination,
	}, nil
}

func (c *Client) Hangup(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.commandOK(ctx, "uuid_kill "+commandWord(callID))
}

func (c *Client) Hold(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.commandOK(ctx, "uuid_hold "+commandWord(callID))
}

func (c *Client) Unhold(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.commandOK(ctx, "uuid_unhold "+commandWord(callID))
}

func (c *Client) Break(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.commandOK(ctx, "uuid_break "+commandWords(callID, "all"))
}

func (c *Client) Transfer(ctx context.Context, req TransferRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	args := []string{req.CallID, req.Destination}
	if req.Dialplan != "" {
		args = append(args, req.Dialplan)
	}
	if req.Context != "" {
		args = append(args, req.Context)
	}
	return c.commandOK(ctx, "uuid_transfer "+commandWords(args...))
}

func (c *Client) Record(ctx context.Context, req RecordRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	action := req.Action
	if action == "" {
		action = "start"
	}
	return c.commandOK(ctx, "uuid_record "+commandWords(req.CallID, action, req.Path))
}
