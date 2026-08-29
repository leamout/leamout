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

	endpoint := req.Endpoint
	if req.CallerID != "" {
		callerIDVar := "origination_caller_id_number=" + commandWord(req.CallerID)
		if prefix := formatVariables(req.Variables); prefix != "" {
			endpoint = strings.TrimSuffix(prefix, "}") + "," + callerIDVar + "}" + endpoint
		} else {
			endpoint = "{" + callerIDVar + "}" + endpoint
		}
	} else if prefix := formatVariables(req.Variables); prefix != "" {
		endpoint = prefix + endpoint
	}

	// Once the B-leg answers, keep the channel parked under ESL control instead
	// of sending it through a user-facing dialplan destination.
	command := "originate " + commandWords(endpoint, "&park()")

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

// Answer answers an incoming call.
func (c *Client) Answer(ctx context.Context, callID string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	return c.commandOK(ctx, "uuid_answer "+commandWord(callID))
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

// PlayAudio plays an audio file to a call without blocking the ESL client.
func (c *Client) PlayAudio(ctx context.Context, callID, filePath string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	filePath, err = requiredArgument("audio file path", filePath)
	if err != nil {
		return err
	}
	return c.commandOK(ctx, "uuid_broadcast "+commandWords(callID, filePath, "aleg"))
}

// StopAudio stops media playback on both legs of a call.
func (c *Client) StopAudio(ctx context.Context, callID string) error {
	return c.Break(ctx, callID)
}

// SendDTMF sends DTMF digits to a call.
func (c *Client) SendDTMF(ctx context.Context, callID, digits string) error {
	callID, err := requiredArgument("call ID", callID)
	if err != nil {
		return err
	}
	digits, err = requiredArgument("DTMF digits", digits)
	if err != nil {
		return err
	}
	return c.commandOK(ctx, "uuid_send_dtmf "+commandWords(callID, digits))
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
