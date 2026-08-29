package calls

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/leamout/leamout/internal/integrations/freeswitch"
)

// FreeSWITCHController adapts the FreeSWITCH client to the calls.Controller interface.
// It translates resolved domain-level call operations into FreeSWITCH-specific commands.
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

// Originate initiates an outbound call using a route already resolved by Leamout.
func (c *FreeSWITCHController) Originate(ctx context.Context, req OriginateRequest) (string, error) {
	endpoint, err := freeSWITCHEndpoint(req)
	if err != nil {
		return "", err
	}

	call, err := c.client.Originate(ctx, freeswitch.OriginateRequest{
		Endpoint:    endpoint,
		Destination: req.Destination,
		CallerID:    req.CallerID,
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

func freeSWITCHEndpoint(req OriginateRequest) (string, error) {
	host := strings.TrimSpace(req.Host)
	if host == "" {
		return "", fmt.Errorf("resolved route host is required")
	}
	if req.Port < 1 || req.Port > 65535 {
		return "", fmt.Errorf("resolved route port is invalid: %d", req.Port)
	}

	transport := strings.ToLower(strings.TrimSpace(req.Transport))
	switch transport {
	case "udp", "tcp", "tls":
	default:
		return "", fmt.Errorf("resolved route transport is invalid: %q", req.Transport)
	}

	destination := strings.TrimSpace(req.Destination)
	if destination == "" {
		return "", fmt.Errorf("resolved route destination is required")
	}

	target := net.JoinHostPort(host, strconv.Itoa(int(req.Port)))
	return fmt.Sprintf("sofia/external/%s@%s;transport=%s", destination, target, transport), nil
}

func (c *FreeSWITCHController) Answer(ctx context.Context, callID string) error {
	if err := c.client.Answer(ctx, callID); err != nil {
		return fmt.Errorf("answer call: %w", err)
	}
	return nil
}

func (c *FreeSWITCHController) Hangup(ctx context.Context, callID string) error {
	if err := c.client.Hangup(ctx, callID); err != nil {
		return fmt.Errorf("hangup call: %w", err)
	}
	return nil
}

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

func (c *FreeSWITCHController) Hold(ctx context.Context, callID string) error {
	if err := c.client.Hold(ctx, callID); err != nil {
		return fmt.Errorf("hold call: %w", err)
	}
	return nil
}

func (c *FreeSWITCHController) Unhold(ctx context.Context, callID string) error {
	if err := c.client.Unhold(ctx, callID); err != nil {
		return fmt.Errorf("unhold call: %w", err)
	}
	return nil
}

func (c *FreeSWITCHController) Play(ctx context.Context, callID, path string) error {
	if err := c.client.PlayAudio(ctx, callID, path); err != nil {
		return fmt.Errorf("play audio: %w", err)
	}
	return nil
}

func (c *FreeSWITCHController) Stop(ctx context.Context, callID string) error {
	if err := c.client.Break(ctx, callID); err != nil {
		return fmt.Errorf("stop audio: %w", err)
	}
	return nil
}

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

func (c *FreeSWITCHController) DTMF(ctx context.Context, callID, digits string) error {
	if err := c.client.SendDTMF(ctx, callID, digits); err != nil {
		return fmt.Errorf("send DTMF: %w", err)
	}
	return nil
}
