package numbers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	redisintegration "github.com/leamout/leamout/internal/integrations/redis"
)

const managedNumberSelectionTTL = 10 * time.Minute

type RedisSelectionStore struct {
	redis *redisintegration.Client
	ttl   time.Duration
}

func NewRedisSelectionStore(redis *redisintegration.Client) *RedisSelectionStore {
	return &RedisSelectionStore{redis: redis, ttl: managedNumberSelectionTTL}
}

func (s *RedisSelectionStore) Save(
	ctx context.Context,
	organizationID uuid.UUID,
	candidate ManagedNumberCandidate,
) (string, error) {
	if s == nil || s.redis == nil {
		return "", fmt.Errorf("managed number selection store is unavailable")
	}
	if organizationID == uuid.Nil {
		return "", fmt.Errorf("organization id is required")
	}

	selectionID := "sel_" + uuid.NewString()
	payload, err := json.Marshal(candidate)
	if err != nil {
		return "", fmt.Errorf("encode managed number selection: %w", err)
	}
	if err := s.redis.Set(ctx, managedNumberSelectionKey(organizationID, selectionID), payload, s.ttl); err != nil {
		return "", fmt.Errorf("store managed number selection: %w", err)
	}
	return selectionID, nil
}

func managedNumberSelectionKey(organizationID uuid.UUID, selectionID string) string {
	return "telecom:numbers:selection:" + organizationID.String() + ":" + selectionID
}
