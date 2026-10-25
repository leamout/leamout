package calling

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	redisintegration "github.com/leamout/leamout/server/internal/integrations/redis"
	redisv9 "github.com/redis/go-redis/v9"
)

type ChannelStore struct {
	client *redisintegration.Client
}

func NewChannelStore(client *redisintegration.Client) *ChannelStore {
	if client == nil {
		panic("calling: Redis client is required")
	}
	return &ChannelStore{
		client: client,
	}
}

func (s *ChannelStore) Bind(ctx context.Context, callID uuid.UUID, channelID string) error {
	channelID = strings.TrimSpace(channelID)
	if callID == uuid.Nil || channelID == "" {
		return fmt.Errorf("call id and channel id are required")
	}

	if err := s.client.Set(ctx, channelKey(callID), channelID, 24*time.Hour); err != nil {
		return err
	}
	if err := s.client.Set(ctx, channelCallKey(channelID), callID.String(), 24*time.Hour); err != nil {
		_ = s.client.Delete(ctx, channelKey(callID))
		return err
	}
	return nil
}

func (s *ChannelStore) Get(ctx context.Context, callID uuid.UUID) (string, error) {
	if callID == uuid.Nil {
		return "", fmt.Errorf("call id is required")
	}

	channelID, err := s.client.Get(ctx, channelKey(callID))
	if errors.Is(err, redisv9.Nil) {
		return "", ErrChannelUnavailable
	}
	if err != nil {
		return "", err
	}

	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return "", ErrChannelUnavailable
	}
	return channelID, nil
}

func (s *ChannelStore) GetCallID(ctx context.Context, channelID string) (uuid.UUID, error) {
	channelID = strings.TrimSpace(channelID)
	if channelID == "" {
		return uuid.Nil, fmt.Errorf("channel id is required")
	}

	value, err := s.client.Get(ctx, channelCallKey(channelID))
	if errors.Is(err, redisv9.Nil) {
		return uuid.Nil, ErrChannelUnavailable
	}
	if err != nil {
		return uuid.Nil, err
	}

	callID, err := uuid.Parse(strings.TrimSpace(value))
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid call id stored for channel: %w", err)
	}
	return callID, nil
}

func (s *ChannelStore) Delete(ctx context.Context, callID uuid.UUID) error {
	if callID == uuid.Nil {
		return fmt.Errorf("call id is required")
	}

	channelID, err := s.client.Get(ctx, channelKey(callID))
	if err != nil && !errors.Is(err, redisv9.Nil) {
		return err
	}

	keys := []string{
		channelKey(callID),
	}
	if channelID = strings.TrimSpace(channelID); channelID != "" {
		keys = append(keys, channelCallKey(channelID))
	}
	return s.client.Delete(ctx, keys...)
}

func channelKey(callID uuid.UUID) string {
	return "telecom:calls:channel:" + callID.String()
}

func channelCallKey(channelID string) string {
	return "telecom:calls:channel-call:" + channelID
}
