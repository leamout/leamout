package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redisv9 "github.com/redis/go-redis/v9"
)

// Get returns the raw value stored at key.
// A missing key is returned as redis.Nil so callers can distinguish a cache miss.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	if err := c.validateKey(key); err != nil {
		return "", err
	}

	return c.client.Get(ctx, key).Result()
}

// Set stores value at key for ttl. A zero ttl stores the value without expiry.
func (c *Client) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if err := c.validateKey(key); err != nil {
		return err
	}

	return c.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes one or more keys.
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	if err := c.validate(); err != nil {
		return err
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

	count, err := c.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetJSON retrieves a JSON-encoded value and unmarshals it into dst.
func (c *Client) GetJSON(ctx context.Context, key string, dst any) error {
	value, err := c.Get(ctx, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(value), dst); err != nil {
		return fmt.Errorf("decode redis value: %w", err)
	}

	return nil
}

// SetJSON JSON-encodes value and stores it at key for ttl.
func (c *Client) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode redis value: %w", err)
	}

	return c.Set(ctx, key, encoded, ttl)
}

func (c *Client) validate() error {
	if c == nil || c.client == nil {
		return fmt.Errorf("redis client is nil")
	}

	return nil
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

var _ = redisv9.Nil
