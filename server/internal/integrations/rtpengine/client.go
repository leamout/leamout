package rtpengine

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	address string
	timeout time.Duration
	dialer  net.Dialer
}

func New(cfg Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Client{
		address: cfg.Address,
		timeout: cfg.CommandTimeout,
		dialer:  net.Dialer{Timeout: cfg.ConnectTimeout},
	}, nil
}

func (c *Client) do(ctx context.Context, command Command, params map[string]any) (Response, error) {
	if ctx == nil {
		return Response{}, fmt.Errorf("RTPEngine context is required")
	}
	if strings.TrimSpace(string(command)) == "" {
		return Response{}, fmt.Errorf("RTPEngine command is required")
	}

	conn, err := c.dialer.DialContext(ctx, "udp", c.address)
	if err != nil {
		return Response{}, fmt.Errorf("dial RTPEngine: %w", err)
	}
	defer func() { _ = conn.Close() }()

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(c.timeout)
	}
	if err := conn.SetDeadline(deadline); err != nil {
		return Response{}, fmt.Errorf("set RTPEngine deadline: %w", err)
	}

	payload := make(map[string]any, len(params)+1)
	for key, value := range params {
		payload[key] = value
	}
	payload["command"] = string(command)

	body, err := bencode(payload)
	if err != nil {
		return Response{}, fmt.Errorf("encode RTPEngine request: %w", err)
	}

	cookie, err := newCookie()
	if err != nil {
		return Response{}, fmt.Errorf("generate RTPEngine cookie: %w", err)
	}

	packet := append([]byte(cookie+" "), body...)
	if _, err := conn.Write(packet); err != nil {
		return Response{}, fmt.Errorf("write RTPEngine request: %w", err)
	}

	buffer := make([]byte, 1<<20)
	n, err := conn.Read(buffer)
	if err != nil {
		return Response{}, fmt.Errorf("read RTPEngine response: %w", err)
	}

	responseCookie, responseBody, ok := strings.Cut(string(buffer[:n]), " ")
	if !ok || responseCookie != cookie {
		return Response{}, fmt.Errorf("invalid RTPEngine response cookie")
	}

	values, err := bdecodeMap([]byte(responseBody))
	if err != nil {
		return Response{}, fmt.Errorf("decode RTPEngine response: %w", err)
	}

	response := Response{
		Result: stringValue(values["result"]),
		Error:  stringValue(values["error-reason"]),
		Data:   values,
	}
	if !response.OK() {
		return response, fmt.Errorf("RTPEngine %s failed: %s", command, response.Error)
	}

	return response, nil
}

func newCookie() (string, error) {
	var value [12]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", value), nil
}

func bencode(value any) ([]byte, error) {
	var buf bytes.Buffer
	if err := encodeValue(&buf, value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeValue(buf *bytes.Buffer, value any) error {
	switch value := value.(type) {
	case string:
		fmt.Fprintf(buf, "%d:%s", len(value), value)
	case []byte:
		fmt.Fprintf(buf, "%d:", len(value))
		_, _ = buf.Write(value)
	case int:
		fmt.Fprintf(buf, "i%de", value)
	case int64:
		fmt.Fprintf(buf, "i%de", value)
	case bool:
		if value {
			buf.WriteString("i1e")
		} else {
			buf.WriteString("i0e")
		}
	case []string:
		buf.WriteByte('l')
		for _, item := range value {
			if err := encodeValue(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte('e')
	case []any:
		buf.WriteByte('l')
		for _, item := range value {
			if err := encodeValue(buf, item); err != nil {
				return err
			}
		}
		buf.WriteByte('e')
	case map[string]any:
		buf.WriteByte('d')
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sortStrings(keys)
		for _, key := range keys {
			if err := encodeValue(buf, key); err != nil {
				return err
			}
			if err := encodeValue(buf, value[key]); err != nil {
				return err
			}
		}
		buf.WriteByte('e')
	default:
		return fmt.Errorf("unsupported bencode type %T", value)
	}
	return nil
}

func bdecodeMap(data []byte) (map[string]any, error) {
	value, offset, err := decodeValue(data, 0)
	if err != nil {
		return nil, err
	}
	if offset != len(data) {
		return nil, fmt.Errorf("trailing data after bencode value")
	}
	result, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("RTPEngine response is not a dictionary")
	}
	return result, nil
}

func decodeValue(data []byte, offset int) (any, int, error) {
	if offset >= len(data) {
		return nil, offset, fmt.Errorf("truncated bencode value")
	}

	switch data[offset] {
	case 'd':
		offset++
		result := make(map[string]any)
		for offset < len(data) && data[offset] != 'e' {
			key, next, err := decodeValue(data, offset)
			if err != nil {
				return nil, offset, err
			}
			keyString, ok := key.(string)
			if !ok {
				return nil, offset, fmt.Errorf("bencode dictionary key is not a string")
			}
			value, next, err := decodeValue(data, next)
			if err != nil {
				return nil, offset, err
			}
			result[keyString] = value
			offset = next
		}
		if offset >= len(data) {
			return nil, offset, fmt.Errorf("unterminated bencode dictionary")
		}
		return result, offset + 1, nil
	case 'l':
		offset++
		var result []any
		for offset < len(data) && data[offset] != 'e' {
			value, next, err := decodeValue(data, offset)
			if err != nil {
				return nil, offset, err
			}
			result = append(result, value)
			offset = next
		}
		if offset >= len(data) {
			return nil, offset, fmt.Errorf("unterminated bencode list")
		}
		return result, offset + 1, nil
	case 'i':
		end := bytes.IndexByte(data[offset+1:], 'e')
		if end < 0 {
			return nil, offset, fmt.Errorf("unterminated bencode integer")
		}
		end += offset + 1
		value, err := strconv.ParseInt(string(data[offset+1:end]), 10, 64)
		if err != nil {
			return nil, offset, fmt.Errorf("invalid bencode integer: %w", err)
		}
		return value, end + 1, nil
	default:
		colon := bytes.IndexByte(data[offset:], ':')
		if colon < 1 {
			return nil, offset, fmt.Errorf("invalid bencode string")
		}
		colon += offset
		length, err := strconv.Atoi(string(data[offset:colon]))
		if err != nil || length < 0 {
			return nil, offset, fmt.Errorf("invalid bencode string length")
		}
		start := colon + 1
		end := start + length
		if end > len(data) {
			return nil, offset, fmt.Errorf("truncated bencode string")
		}
		return string(data[start:end]), end, nil
	}
}

func stringValue(value any) string {
	result, ok := value.(string)
	if !ok {
		return ""
	}
	return result
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
