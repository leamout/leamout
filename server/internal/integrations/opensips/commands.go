package opensips

import (
	"context"
	"fmt"
)

const (
	CommandPing       = "ping"
	CommandReload     = "reload"
	CommandSetFlag    = "set_flag"
	CommandResetFlag  = "reset_flag"
)

func (c *Client) Ping(ctx context.Context) (Response, error) {
	return c.Command(ctx, Command{Name: CommandPing})
}

func (c *Client) Reload(ctx context.Context) (Response, error) {
	return c.Command(ctx, Command{Name: CommandReload})
}

func (c *Client) SetFlag(ctx context.Context, name string) (Response, error) {
	if name == "" {
		return Response{}, fmt.Errorf("OpenSIPS flag name is required")
	}
	return c.Command(ctx, Command{Name: CommandSetFlag, Params: map[string]string{"name": name}})
}

func (c *Client) ResetFlag(ctx context.Context, name string) (Response, error) {
	if name == "" {
		return Response{}, fmt.Errorf("OpenSIPS flag name is required")
	}
	return c.Command(ctx, Command{Name: CommandResetFlag, Params: map[string]string{"name": name}})
}
