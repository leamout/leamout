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

// MediaController defines the contract for FreeSWITCH media server operations.
// This interface enables testability and allows swapping implementations.
type MediaController interface {
	Connect(ctx context.Context) error
	Close() error
	HealthCheck(ctx context.Context) error
	Command(ctx context.Context, command string) (Reply, error)
	BGAPI(ctx context.Context, command string) (Job, error)
	Subscribe(ctx context.Context, format EventFormat, events []string, handler EventHandler) error
}

type Client struct {
	address        string
	password       string
	connectTimeout time.Duration
	commandTimeout time.Duration
	dialer         *net.Dialer

	mu      sync.Mutex
	conn    net.Conn
	reader  *bufio.Reader
	done    chan struct{}
	readErr chan error
	replyCh chan Frame

	writeMu   sync.Mutex
	commandMu sync.Mutex

	handlersMu sync.RWMutex
	handlers   []EventHandler
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
	if c.conn != nil {
		c.mu.Unlock()
		return nil
	}
	c.mu.Unlock()

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

	c.mu.Lock()
	if c.conn != nil {
		c.mu.Unlock()
		_ = conn.Close()
		return nil
	}
	c.conn = conn
	c.reader = reader
	c.done = make(chan struct{})
	c.readErr = make(chan error, 1)
	c.replyCh = make(chan Frame)
	c.mu.Unlock()

	go c.readLoop(conn, reader)
	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	if c.conn == nil {
		c.mu.Unlock()
		return nil
	}
	conn := c.conn
	c.conn = nil
	c.reader = nil
	done := c.done
	c.done = nil
	c.mu.Unlock()

	err := conn.Close()
	if done != nil {
		<-done
	}
	return err
}

func (c *Client) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	reply, err := c.Command(ctx, "status")
	if err != nil {
		return fmt.Errorf("FreeSWITCH health check failed: %w", err)
	}
	if !freeSWITCHStatusUp(reply) {
		return fmt.Errorf("FreeSWITCH is not running properly")
	}
	return nil
}

func freeSWITCHStatusUp(reply Reply) bool {
	for _, value := range []string{reply.Body, reply.Text} {
		status := strings.ToUpper(strings.TrimSpace(value))
		if status == "UP" ||
			strings.HasPrefix(status, "UP ") ||
			strings.HasPrefix(status, "+OK UP") {
			return true
		}
	}

	return false
}

func (c *Client) command(ctx context.Context, command string) (Frame, error) {
	if ctx == nil {
		return Frame{}, fmt.Errorf("FreeSWITCH context is required")
	}
	command = strings.TrimSpace(command)
	if command == "" {
		return Frame{}, fmt.Errorf("FreeSWITCH command is required")
	}

	c.commandMu.Lock()
	defer c.commandMu.Unlock()

	c.mu.Lock()
	conn := c.conn
	replyCh := c.replyCh
	done := c.done
	readErr := c.readErr
	c.mu.Unlock()
	if conn == nil || replyCh == nil || done == nil || readErr == nil {
		return Frame{}, fmt.Errorf("FreeSWITCH client is not connected")
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.commandTimeout)
	}
	if err := conn.SetWriteDeadline(deadline); err != nil {
		return Frame{}, fmt.Errorf("set FreeSWITCH command deadline: %w", err)
	}

	c.writeMu.Lock()
	_, err := writeCommand(conn, command)
	c.writeMu.Unlock()
	if err != nil {
		return Frame{}, fmt.Errorf("write FreeSWITCH command: %w", err)
	}

	select {
	case frame := <-replyCh:
		return frame, nil
	case err := <-readErr:
		return Frame{}, fmt.Errorf("read FreeSWITCH response: %w", err)
	case <-done:
		return Frame{}, fmt.Errorf("FreeSWITCH connection closed")
	case <-ctx.Done():
		return Frame{}, fmt.Errorf("FreeSWITCH command: %w", ctx.Err())
	}
}

func (c *Client) readLoop(conn net.Conn, reader *bufio.Reader) {
	defer func() {
		c.mu.Lock()
		if c.conn == conn {
			c.conn = nil
			c.reader = nil
			if c.readErr != nil {
				select {
				case c.readErr <- io.EOF:
				default:
				}
			}
			if c.done != nil {
				close(c.done)
			}
		}
		c.mu.Unlock()
	}()

	for {
		frame, err := readFrame(reader)
		if err != nil {
			c.mu.Lock()
			if c.readErr != nil {
				select {
				case c.readErr <- err:
				default:
				}
			}
			c.mu.Unlock()
			return
		}

		switch frame.ContentType {
		case ContentTypeCommandReply, ContentTypeAPIResponse:
			c.replyCh <- frame
		case ContentTypeEventPlain:
			c.dispatchEvent(frame)
		}
	}
}

func (c *Client) dispatchEvent(frame Frame) {
	event := Event{
		Headers: frame.Headers,
		Body:    frame.Body,
		Name:    frame.Header("Event-Name"),
	}

	c.handlersMu.RLock()
	handlers := append([]EventHandler(nil), c.handlers...)
	c.handlersMu.RUnlock()

	for _, handler := range handlers {
		go func(h EventHandler) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = h(ctx, event)
		}(handler)
	}
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
		ContentType: headers["Content-Type"],
		Headers:     headers,
		Body:        string(body),
	}, nil
}
