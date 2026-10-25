package calling

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/integrations/freeswitch"
)

type Controller struct {
	client *freeswitch.Client
}

func NewController(client *freeswitch.Client) *Controller {
	if client == nil {
		panic("calling: FreeSWITCH client is required")
	}
	return &Controller{client: client}
}
func (c *Controller) Answer(ctx context.Context, channelID string) error {
	if err := c.client.Answer(ctx, channelID); err != nil {
		return fmt.Errorf("answer call: %w", err)
	}
	return nil
}

func (c *Controller) Hangup(ctx context.Context, channelID string) error {
	if err := c.client.Hangup(ctx, channelID); err != nil {
		return fmt.Errorf("hangup call: %w", err)
	}
	return nil
}

func (c *Controller) SetMaxDuration(
	ctx context.Context,
	channelID string,
	seconds int32,
) error {
	if seconds <= 0 {
		return fmt.Errorf("maximum call duration must be positive")
	}
	value := fmt.Sprintf(
		"sched_hangup +%d ALLOTTED_TIMEOUT",
		seconds,
	)
	if err := c.client.SetVariable(
		ctx,
		channelID,
		"execute_on_answer",
		value,
	); err != nil {
		return fmt.Errorf("set maximum call duration: %w", err)
	}

	return nil
}

func (c *Controller) Transfer(
	ctx context.Context,
	channelID string,
	req TransferRequest,
) error {
	if err := c.client.Transfer(ctx, freeswitch.TransferRequest{
		CallID:      channelID,
		Destination: req.Destination,
		Dialplan:    req.Dialplan,
		Context:     req.Context,
	}); err != nil {
		return fmt.Errorf("transfer call: %w", err)
	}
	return nil
}

func (c *Controller) Hold(ctx context.Context, channelID string) error {
	if err := c.client.Hold(ctx, channelID); err != nil {
		return fmt.Errorf("hold call: %w", err)
	}
	return nil
}

func (c *Controller) Resume(ctx context.Context, channelID string) error {
	if err := c.client.Unhold(ctx, channelID); err != nil {
		return fmt.Errorf("resume call: %w", err)
	}
	return nil
}

func (c *Controller) PlayAudio(ctx context.Context, channelID, path string) error {
	if err := c.client.PlayAudio(ctx, channelID, path); err != nil {
		return fmt.Errorf("play audio: %w", err)
	}
	return nil
}

func (c *Controller) StopPlayback(ctx context.Context, channelID string) error {
	if err := c.client.StopAudio(ctx, channelID); err != nil {
		return fmt.Errorf("stop audio: %w", err)
	}
	return nil
}

func (c *Controller) Record(
	ctx context.Context,
	channelID string,
	req RecordRequest,
) error {
	if err := c.client.Record(ctx, freeswitch.RecordRequest{
		CallID: channelID,
		Path:   req.Path,
		Action: req.Action,
	}); err != nil {
		return fmt.Errorf("record call: %w", err)
	}
	return nil
}

func (c *Controller) SendDTMF(ctx context.Context, channelID, digits string) error {
	if err := c.client.SendDTMF(ctx, channelID, digits); err != nil {
		return fmt.Errorf("send DTMF: %w", err)
	}
	return nil
}

func (c *Controller) SetCallID(
	ctx context.Context,
	channelID string,
	callID uuid.UUID,
) error {
	if callID == uuid.Nil {
		return fmt.Errorf("call id is required")
	}
	if err := c.client.SetVariable(ctx, channelID, leamoutCallIDVar, callID.String()); err != nil {
		return fmt.Errorf("set call id: %w", err)
	}
	return nil
}

func (c *Controller) ChannelExists(ctx context.Context, channelID string) (bool, error) {
	return c.client.ChannelExists(ctx, channelID)
}
