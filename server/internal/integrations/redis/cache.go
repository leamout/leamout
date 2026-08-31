package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Get returns the value stored at key. A missing key returns redis.Nil.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if err := c.validateKey(key); err != nil {
		return "", err
	}
	if ctx == nil {
		return "", fmt.Errorf("redis context is nil")
	}

	return c.client.Get(ctx, key).Result()
}

// Set stores value at key. A zero TTL stores the value without expiry.
func (c *Client) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if err := c.validateKey(key); err != nil {
		return err
	}
	if ctx == nil {
		return fmt.Errorf("redis context is nil")
	}

	return c.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes keys. It is a no-op when no keys are provided.
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if err := c.validate(); err != nil {
		return err
	}
	if ctx == nil {
		return fmt.Errorf("redis context is nil")
	}
	if len(keys) == 0 {
		return nil
	}

	for _, key := range keys {
		if key == "" {
			return fmt.Errorf("redis key is required")
		}
	}

	return c.client.Del(ctx, keys...).Err()
}

// Exists reports whether key exists.
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	if err := c.validateKey(key); err != nil {
		return false, err
	}
	if ctx == nil {
		return false, fmt.Errorf("redis context is nil")
	}

	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// AllowFixedWindow atomically increments a bounded, expiring counter and
// reports whether the request remains within limit. The Lua script ensures
// every API replica shares the same decision and that counters cannot persist
// without an expiry after a process failure.
func (c *Client) AllowFixedWindow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	if err := c.validateKey(key); err != nil {
		return false, err
	}
	if ctx == nil {
		return false, fmt.Errorf("redis context is nil")
	}
	if limit <= 0 {
		return false, fmt.Errorf("rate limit must be positive")
	}
	if window <= 0 {
		return false, fmt.Errorf("rate limit window must be positive")
	}

	const script = `
local current = redis.call('INCR', KEYS[1])
if current == 1 then
  redis.call('PEXPIRE', KEYS[1], ARGV[1])
end
if current <= tonumber(ARGV[2]) then
  return 1
end
return 0
`
	allowed, err := c.client.Eval(ctx, script, []string{key}, window.Milliseconds(), limit).Bool()
	if err != nil {
		return false, fmt.Errorf("apply Redis fixed-window rate limit: %w", err)
	}
	return allowed, nil
}

// GetJSON retrieves a JSON value into dst.
func (c *Client) GetJSON(ctx context.Context, key string, dst any) error {
	if dst == nil {
		return fmt.Errorf("redis destination is nil")
	}

	value, err := c.Get(ctx, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(value), dst); err != nil {
		return fmt.Errorf("decode redis value: %w", err)
	}

	return nil
}

// SetJSON JSON-encodes value and stores it at key.
func (c *Client) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode redis value: %w", err)
	}

	return c.Set(ctx, key, encoded, ttl)
}

func (c *Client) validateKey(key string) error {
	if err := c.validate(); err != nil {
		return err
	}
	if key == "" {
		return fmt.Errorf("redis key is required")
	}

	return nil
}
