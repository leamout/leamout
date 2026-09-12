package usage

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

type store interface {
	CreateEvent(context.Context, uuid.UUID, RecordInput) (Event, error)
	GetByIdempotencyKey(context.Context, uuid.UUID, string) (Event, error)
}

// Service owns validation and organization-scoped idempotency for authoritative
// usage observations. Billing decisions are made separately from recording use.
type Service struct {
	repo store
}

func NewService(repo store) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(
	ctx context.Context,
	organizationID uuid.UUID,
	input RecordInput,
) (RecordResult, error) {
	normalized, err := normalizeRecordInput(input)
	if err != nil || organizationID == uuid.Nil {
		return RecordResult{}, ErrInvalidEvent
	}

	event, err := s.repo.CreateEvent(ctx, organizationID, normalized)
	if err == nil {
		return RecordResult{Event: event}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return RecordResult{}, err
	}

	existing, err := s.repo.GetByIdempotencyKey(
		ctx,
		organizationID,
		normalized.IdempotencyKey,
	)
	if err != nil {
		return RecordResult{}, err
	}
	if !sameEvent(existing, normalized) {
		return RecordResult{}, ErrEventConflict
	}

	return RecordResult{Event: existing, Replayed: true}, nil
}

func normalizeRecordInput(input RecordInput) (RecordInput, error) {
	input.SourceType = strings.TrimSpace(input.SourceType)
	input.SourceID = strings.TrimSpace(input.SourceID)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)

	if input.MeterID == uuid.Nil || input.Quantity <= 0 || input.OccurredAt.IsZero() ||
		input.SourceID == "" || input.IdempotencyKey == "" ||
		!sourceTypePattern.MatchString(input.SourceType) {
		return RecordInput{}, ErrInvalidEvent
	}

	if len(input.Dimensions) == 0 {
		input.Dimensions = json.RawMessage(`{}`)
	}
	var dimensions map[string]any
	if err := json.Unmarshal(input.Dimensions, &dimensions); err != nil || dimensions == nil {
		return RecordInput{}, ErrInvalidEvent
	}

	input.OccurredAt = input.OccurredAt.UTC()
	return input, nil
}

func sameEvent(existing Event, input RecordInput) bool {
	if existing.MeterID != input.MeterID || existing.Quantity != input.Quantity ||
		existing.SourceType != input.SourceType || existing.SourceID != input.SourceID ||
		!existing.OccurredAt.Equal(input.OccurredAt) {
		return false
	}

	var existingDimensions any
	var inputDimensions any
	if json.Unmarshal(existing.Dimensions, &existingDimensions) != nil ||
		json.Unmarshal(input.Dimensions, &inputDimensions) != nil {
		return false
	}

	return reflect.DeepEqual(existingDimensions, inputDimensions)
}
