package opensips

import (
	"context"
	"fmt"
	"net/netip"
	"strconv"
	"strings"
)

const (
	CommandReloadSubscriberCache = "reload_subscriber_cache"
	CommandFlushSubscriberCache  = "flush_subscriber_cache"
	CommandAddRoute              = "add_route"
	CommandRemoveRoute           = "remove_route"
	CommandListRoutes            = "list_routes"
	CommandListDialogs           = "list_dialogs"
	CommandGetDialog             = "get_dialog"
	CommandTerminateDialog       = "terminate_dialog"
	CommandBlockIP               = "block_ip"
	CommandUnblockIP             = "unblock_ip"
	CommandIsBlocked             = "is_blocked"
	CommandListBlockedIPs        = "list_blocked_ips"
	CommandGetStatistics         = "get_statistics"
)

func (c *Client) ReloadSubscriberCache(ctx context.Context, req SubscriberCacheRequest) (Response, error) {
	params := map[string]string{}
	if username := strings.TrimSpace(req.Username); username != "" {
		params["username"] = username
	}
	if domain := strings.TrimSpace(req.Domain); domain != "" {
		params["domain"] = domain
	}
	return c.Command(ctx, Command{Name: CommandReloadSubscriberCache, Params: params})
}

func (c *Client) FlushSubscriberCache(ctx context.Context) (Response, error) {
	return c.Command(ctx, Command{Name: CommandFlushSubscriberCache})
}

func (c *Client) AddRoute(ctx context.Context, route Route) (Response, error) {
	if err := validateRoute(route); err != nil {
		return Response{}, err
	}
	return c.Command(ctx, Command{Name: CommandAddRoute, Params: routeParams(route)})
}

func (c *Client) RemoveRoute(ctx context.Context, routeID string) (Response, error) {
	routeID = strings.TrimSpace(routeID)
	if routeID == "" {
		return Response{}, fmt.Errorf("OpenSIPS route ID is required")
	}
	return c.Command(ctx, Command{Name: CommandRemoveRoute, Params: map[string]string{"id": routeID}})
}

func (c *Client) ListRoutes(ctx context.Context) (Response, error) {
	return c.Command(ctx, Command{Name: CommandListRoutes})
}

func (c *Client) ListDialogs(ctx context.Context) (Response, error) {
	return c.Command(ctx, Command{Name: CommandListDialogs})
}

func (c *Client) GetDialog(ctx context.Context, dialogID string) (Response, error) {
	dialogID = strings.TrimSpace(dialogID)
	if dialogID == "" {
		return Response{}, fmt.Errorf("OpenSIPS dialog ID is required")
	}
	return c.Command(ctx, Command{Name: CommandGetDialog, Params: map[string]string{"id": dialogID}})
}

func (c *Client) TerminateDialog(ctx context.Context, dialogID string) (Response, error) {
	dialogID = strings.TrimSpace(dialogID)
	if dialogID == "" {
		return Response{}, fmt.Errorf("OpenSIPS dialog ID is required")
	}
	return c.Command(ctx, Command{Name: CommandTerminateDialog, Params: map[string]string{"id": dialogID}})
}

func (c *Client) BlockIP(ctx context.Context, entry AccessEntry) (Response, error) {
	return c.accessCommand(ctx, CommandBlockIP, entry)
}

func (c *Client) UnblockIP(ctx context.Context, address string) (Response, error) {
	address, err := normalizeIP(address)
	if err != nil {
		return Response{}, err
	}
	return c.Command(ctx, Command{Name: CommandUnblockIP, Params: map[string]string{"address": address}})
}

func (c *Client) IsBlocked(ctx context.Context, address string) (Response, error) {
	address, err := normalizeIP(address)
	if err != nil {
		return Response{}, err
	}
	return c.Command(ctx, Command{Name: CommandIsBlocked, Params: map[string]string{"address": address}})
}

func (c *Client) ListBlockedIPs(ctx context.Context) (Response, error) {
	return c.Command(ctx, Command{Name: CommandListBlockedIPs})
}

func (c *Client) GetStatistics(ctx context.Context) (Response, error) {
	return c.Command(ctx, Command{Name: CommandGetStatistics})
}

func (c *Client) GetActiveCalls(ctx context.Context) (Response, error) {
	return c.Command(ctx, Command{Name: CommandGetStatistics, Params: map[string]string{"metric": "active_calls"}})
}

func (c *Client) GetCPS(ctx context.Context) (Response, error) {
	return c.Command(ctx, Command{Name: CommandGetStatistics, Params: map[string]string{"metric": "cps"}})
}

func routeParams(route Route) map[string]string {
	return map[string]string{
		"id":       strings.TrimSpace(route.ID),
		"carrier":  strings.TrimSpace(route.Carrier),
		"prefix":   strings.TrimSpace(route.Prefix),
		"uri":      strings.TrimSpace(route.URI),
		"priority": strconv.Itoa(route.Priority),
		"enabled":  strconv.FormatBool(route.Enabled),
	}
}

func validateRoute(route Route) error {
	if strings.TrimSpace(route.ID) == "" {
		return fmt.Errorf("OpenSIPS route ID is required")
	}
	if strings.TrimSpace(route.Carrier) == "" {
		return fmt.Errorf("OpenSIPS route carrier is required")
	}
	if strings.TrimSpace(route.URI) == "" {
		return fmt.Errorf("OpenSIPS route URI is required")
	}
	if route.Priority < 0 {
		return fmt.Errorf("OpenSIPS route priority must not be negative")
	}
	return nil
}

func (c *Client) accessCommand(ctx context.Context, command string, entry AccessEntry) (Response, error) {
	address, err := normalizeIP(entry.Address)
	if err != nil {
		return Response{}, err
	}
	params := map[string]string{"address": address}
	if reason := strings.TrimSpace(entry.Reason); reason != "" {
		params["reason"] = reason
	}
	if !entry.ExpiresAt.IsZero() {
		params["expires_at"] = entry.ExpiresAt.UTC().Format("2006-01-02T15:04:05Z07:00")
	}
	return c.Command(ctx, Command{Name: command, Params: params})
}

func normalizeIP(value string) (string, error) {
	address, err := netip.ParseAddr(strings.TrimSpace(value))
	if err != nil {
		return "", fmt.Errorf("invalid OpenSIPS IP address %q: %w", value, err)
	}
	return address.String(), nil
}
