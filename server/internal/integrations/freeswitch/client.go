package freeswitch

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Client struct {
	address  string
	password string

	connectTimeout time.Duration
	commandTimeout time.Duration
	dialer         *net.Dialer

	mu     sync.Mutex
	conn   net.Conn
	reader *bufio.Reader
}

func New(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Client{
		address:        cfg.Address,
		password:       cfg.Password,
		connectTimeout: cfg.ConnectTimeout,
		commandTimeout: cfg.CommandTimeout,
		dialer:         &net.Dialer{Timeout: cfg.ConnectTimeout},
	}, nil
}

func (c *Client) Connect(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("FreeSWITCH context is required")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return nil
	}

	conn, err := c.dialer.DialContext(ctx, "tcp", c.address)
	if err != nil {
		return fmt.Errorf("connect to FreeSWITCH: %w", err)
	}

	reader := bufio.NewReader(conn)
	frame, err := readFrame(reader)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("read FreeSWITCH authentication request: %w", err)
	}
	if frame.ContentType != ContentTypeAuthRequest {
		_ = conn.Close()
		return fmt.Errorf("unexpected FreeSWITCH greeting: %q", frame.ContentType)
	}

	if _, err := writeCommand(conn, "auth "+c.password); err != nil {
		_ = conn.Close()
		return fmt.Errorf("write FreeSWITCH authentication command: %w", err)
	}

	frame, err = readFrame(reader)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("read FreeSWITCH authentication response: %w", err)
	}
	if frame.ContentType != ContentTypeCommandReply || !frame.OK() {
		_ = conn.Close()
		return fmt.Errorf("FreeSWITCH authentication failed: %s", frame.ReplyText())
	}

	c.conn = conn
	c.reader = reader
	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	err := c.conn.Close()
	c.conn = nil
	c.reader = nil
	return err
}

func (c *Client) command(ctx context.Context, command string) (Frame, error) {
	if ctx == nil {
		return Frame{}, fmt.Errorf("FreeSWITCH context is required")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil || c.reader == nil {
		return Frame{}, fmt.Errorf("FreeSWITCH client is not connected")
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.commandTimeout)
	}
	if err := c.conn.SetDeadline(deadline); err != nil {
		return Frame{}, fmt.Errorf("set FreeSWITCH command deadline: %w", err)
	}
	defer func() { _ = c.conn.SetDeadline(time.Time{}) }()

	if _, err := writeCommand(c.conn, command); err != nil {
		return Frame{}, fmt.Errorf("write FreeSWITCH command: %w", err)
	}

	frame, err := readFrame(c.reader)
	if err != nil {
		return Frame{}, fmt.Errorf("read FreeSWITCH response: %w", err)
	}

	return frame, nil
}

func writeCommand(conn net.Conn, command string) (int, error) {
	return fmt.Fprintf(conn, "%s\n\n", strings.TrimSpace(command))
}

func readFrame(reader *bufio.Reader) (Frame, error) {
	headers := make(map[string]string)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return Frame{}, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}

		key, value, ok := strings.Cut(line, ": ")
		if !ok {
			return Frame{}, fmt.Errorf("invalid FreeSWITCH header %q", line)
		}
		headers[key] = value
	}

	contentType := headers["Content-Type"]
	length := 0
	if raw := headers["Content-Length"]; raw != "" {
		var err error
		length, err = strconv.Atoi(raw)
		if err != nil || length < 0 {
			return Frame{}, fmt.Errorf("invalid FreeSWITCH content length %q", raw)
		}
	}

	body := make([]byte, length)
	if length > 0 {
		if _, err := io.ReadFull(reader, body); err != nil {
			return Frame{}, fmt.Errorf("read FreeSWITCH frame body: %w", err)
		}
	}

	return Frame{
		ContentType: contentType,
		Headers:     headers,
		Body:        string(body),
	}, nil
}
