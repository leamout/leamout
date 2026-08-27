package calls

import (
	"context"
	"fmt"

	"github.com/leamout/leamout/internal/integrations/freeswitch"
)

// FreeSWITCHController adapts the FreeSWITCH client to the calls.Controller interface.
// It translates domain-level call operations into FreeSWITCH-specific commands.
type FreeSWITCHController struct {
	client *freeswitch.Client
}

var _ Controller = (*FreeSWITCHController)(nil)

// NewFreeSWITCHController creates a new adapter wrapping the FreeSWITCH client.
func NewFreeSWITCHController(client *freeswitch.Client) *FreeSWITCHController {
	if client == nil {
		panic("calls: FreeSWITCH client is required")
	}

	return &FreeSWITCHController{client: client}
}

// Originate initiates an outbound call and returns the SIP call ID (UUID).
func (c *FreeSWITCHController) Originate(ctx context.Context, req CreateCallRequest) (string, error) {
	call, err := c.client.Originate(ctx, freeswitch.OriginateRequest{
		Endpoint:    req.Endpoint,
		Destination: req.To,
		CallerID:    req.From,
		Variables:   req.Variables,
	})
	if err != nil {
		return "", fmt.Errorf("originate call: %w", err)
	}
	if call.UUID == "" {
		return "", fmt.Errorf("FreeSWITCH returned empty call UUID")
	}

	return call.UUID, nil
}

// Answer forces an inbound call into the answered state.
func (c *FreeSWITCHController) Answer(ctx context.Context, callID string) error {
	if err := c.client.Answer(ctx, callID); err != nil {
		return fmt.Errorf("answer call: %w", err)
	}

	return nil
}

// Hangup terminates an active call.
func (c *FreeSWITCHController) Hangup(ctx context.Context, callID string) error {
	if err := c.client.Hangup(ctx, callID); err != nil {
		return fmt.Errorf("hangup call: %w", err)
	}

	return nil
}

// Transfer moves a call to a new destination.
func (c *FreeSWITCHController) Transfer(ctx context.Context, callID string, req TransferRequest) error {
	if err := c.client.Transfer(ctx, freeswitch.TransferRequest{
		CallID:      callID,
		Destination: req.Destination,
		Dialplan:    req.Dialplan,
		Context:     req.Context,
	}); err != nil {
		return fmt.Errorf("transfer call: %w", err)
	}

	return nil
}

// Hold places a call on hold.
func (c *FreeSWITCHController) Hold(ctx context.Context, callID string) error {
	if err := c.client.Hold(ctx, callID); err != nil {
		return fmt.Errorf("hold call: %w", err)
	}

	return nil
}

// Unhold resumes a held call.
func (c *FreeSWITCHController) Unhold(ctx context.Context, callID string) error {
	if err := c.client.Unhold(ctx, callID); err != nil {
		return fmt.Errorf("unhold call: %w", err)
	}

	return nil
}

// Play plays an audio file to the call.
func (c *FreeSWITCHController) Play(ctx context.Context, callID, path string) error {
	if err := c.client.PlayAudio(ctx, callID, path); err != nil {
		return fmt.Errorf("play audio: %w", err)
	}

	return nil
}

// Stop stops any audio playback on the call.
func (c *FreeSWITCHController) Stop(ctx context.Context, callID string) error {
	// Uses the existing Break command which stops all media on the channel.
	if err := c.client.Break(ctx, callID); err != nil {
		return fmt.Errorf("stop audio: %w", err)
	}

	return nil
}

// Record starts or stops recording a call.
func (c *FreeSWITCHController) Record(ctx context.Context, callID string, req RecordRequest) error {
	if err := c.client.Record(ctx, freeswitch.RecordRequest{
		CallID: callID,
		Path:   req.Path,
		Action: req.Action,
	}); err != nil {
		return fmt.Errorf("record call: %w", err)
	}

	return nil
}

// DTMF sends DTMF digits to the call.
func (c *FreeSWITCHController) DTMF(ctx context.Context, callID, digits string) error {
	if err := c.client.SendDTMF(ctx, callID, digits); err != nil {
		return fmt.Errorf("send DTMF: %w", err)
	}

	return nil
}
