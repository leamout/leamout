package metering

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var sourceTypePattern = regexp.MustCompile(`^[a-z0-9]+(?:_[a-z0-9]+)*$`)

// Service owns validation and organization-scoped idempotency for usage
// observations. Whether an observation is billable is decided later from the
// organization's commercial mode and an applicable metered price.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetMeter(ctx context.Context, key string) (Meter, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return Meter{}, ErrInvalidUsageEvent
	}
	return s.repo.GetMeter(ctx, key)
}

func (s *Service) Record(ctx context.Context, organizationID uuid.UUID, input RecordInput) (RecordResult, error) {
	normalized, err := normalizeRecordInput(input)
	if err != nil || organizationID == uuid.Nil {
		return RecordResult{}, ErrInvalidUsageEvent
	}

	event, err := s.repo.CreateUsageEvent(ctx, organizationID, normalized)
	if err == nil {
		return RecordResult{Event: event}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return RecordResult{}, err
	}

	existing, err := s.repo.GetUsageEventByIdempotencyKey(ctx, organizationID, normalized.IdempotencyKey)
	if err != nil {
		return RecordResult{}, err
	}
	if !sameUsage(existing, normalized) {
		return RecordResult{}, ErrUsageEventConflict
	}
	return RecordResult{Event: existing, Replayed: true}, nil
}

func normalizeRecordInput(input RecordInput) (RecordInput, error) {
	input.SourceType = strings.TrimSpace(input.SourceType)
	input.SourceID = strings.TrimSpace(input.SourceID)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.MeterID == uuid.Nil || input.Quantity <= 0 || input.OccurredAt.IsZero() ||
		input.SourceID == "" || input.IdempotencyKey == "" || !sourceTypePattern.MatchString(input.SourceType) {
		return RecordInput{}, ErrInvalidUsageEvent
	}
	if input.SubscriptionID != nil && *input.SubscriptionID == uuid.Nil {
		return RecordInput{}, ErrInvalidUsageEvent
	}
	if len(input.Dimensions) == 0 {
		input.Dimensions = json.RawMessage(`{}`)
	}
	var dimensions map[string]any
	if err := json.Unmarshal(input.Dimensions, &dimensions); err != nil || dimensions == nil {
		return RecordInput{}, ErrInvalidUsageEvent
	}
	input.OccurredAt = input.OccurredAt.UTC()
	return input, nil
}

func sameUsage(existing UsageEvent, input RecordInput) bool {
	if existing.MeterID != input.MeterID || existing.Quantity != input.Quantity ||
		existing.SourceType != input.SourceType || existing.SourceID != input.SourceID ||
		!existing.OccurredAt.Equal(input.OccurredAt) || !sameUUID(existing.SubscriptionID, input.SubscriptionID) {
		return false
	}
	var existingDimensions any
	var inputDimensions any
	if json.Unmarshal(existing.Dimensions, &existingDimensions) != nil || json.Unmarshal(input.Dimensions, &inputDimensions) != nil {
		return false
	}
	return reflect.DeepEqual(existingDimensions, inputDimensions)
}

func sameUUID(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
