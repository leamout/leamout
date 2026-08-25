package rtpengine

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"strings"
	"time"
)

type Client struct { address string; timeout time.Duration; dialer net.Dialer }

func New(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil { return nil, err }
	return &Client{address: cfg.Address, timeout: cfg.CommandTimeout, dialer: net.Dialer{Timeout: cfg.ConnectTimeout}}, nil
}

func (c *Client) do(ctx context.Context, command Command, params map[string]any) (Response, error) {
	if ctx == nil { return Response{}, fmt.Errorf("RTPEngine context is required") }
	conn, err := c.dialer.DialContext(ctx, "udp", c.address)
	if err != nil { return Response{}, fmt.Errorf("dial RTPEngine: %w", err) }
	defer func() { _ = conn.Close() }()
	deadline, ok := ctx.Deadline(); if !ok { deadline = time.Now().Add(c.timeout) }
	if err := conn.SetDeadline(deadline); err != nil { return Response{}, fmt.Errorf("set RTPEngine deadline: %w", err) }
	payload := make(map[string]any, len(params)+1)
	for key, value := range params { payload[key] = value }
	payload["command"] = string(command)
	body, err := bencode(payload); if err != nil { return Response{}, fmt.Errorf("encode RTPEngine request: %w", err) }
	cookie, err := newCookie(); if err != nil { return Response{}, err }
	if _, err := conn.Write(append([]byte(cookie+" "), body...)); err != nil { return Response{}, fmt.Errorf("write RTPEngine request: %w", err) }
	buffer := make([]byte, 1<<20)
	n, err := conn.Read(buffer); if err != nil { return Response{}, fmt.Errorf("read RTPEngine response: %w", err) }
	responseCookie, responseBody, ok := strings.Cut(string(buffer[:n]), " ")
	if !ok || responseCookie != cookie { return Response{}, fmt.Errorf("invalid RTPEngine response cookie") }
	values, err := bdecodeMap([]byte(responseBody)); if err != nil { return Response{}, fmt.Errorf("decode RTPEngine response: %w", err) }
	response := Response{Result: stringValue(values["result"]), Error: stringValue(values["error-reason"]), Data: values}
	if !response.OK() { return response, fmt.Errorf("RTPEngine %s failed: %s", command, response.Error) }
	return response, nil
}

func newCookie() (string, error) { var value [12]byte; if _, err := rand.Read(value[:]); err != nil { return "", err }; return fmt.Sprintf("%x", value), nil }
