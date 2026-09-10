package metering

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type fakeStore struct {
	createEvent UsageEvent
	createErr   error
	existing    UsageEvent
}

func (f *fakeStore) GetMeter(context.Context, string) (Meter, error) {
	return Meter{}, nil
}

func (f *fakeStore) CreateUsageEvent(context.Context, uuid.UUID, RecordInput) (UsageEvent, error) {
	return f.createEvent, f.createErr
}

func (f *fakeStore) GetUsageEventByIdempotencyKey(context.Context, uuid.UUID, string) (UsageEvent, error) {
	return f.existing, nil
}

func TestRecordCreatesUsageEvent(t *testing.T) {
	organizationID := uuid.New()
	meterID := uuid.New()
	occurredAt := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	event := UsageEvent{ID: uuid.New(), OrganizationID: organizationID, MeterID: meterID}
	service := NewService(&fakeStore{createEvent: event})

	result, err := service.Record(t.Context(), organizationID, RecordInput{
		MeterID: meterID, Quantity: 60, SourceType: "voice_call", SourceID: "call-1",
		IdempotencyKey: "voice:call-1", OccurredAt: occurredAt,
	})
	if err != nil {
		t.Fatalf("record usage: %v", err)
	}
	if result.Replayed {
		t.Fatal("new usage must not be marked replayed")
	}
	if result.Event.ID != event.ID {
		t.Fatalf("expected event %s, got %s", event.ID, result.Event.ID)
	}
}

func TestRecordReturnsMatchingReplay(t *testing.T) {
	organizationID := uuid.New()
	meterID := uuid.New()
	occurredAt := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	existing := UsageEvent{
		ID: uuid.New(), OrganizationID: organizationID, MeterID: meterID, Quantity: 60,
		SourceType: "voice_call", SourceID: "call-1", IdempotencyKey: "voice:call-1",
		Dimensions: []byte(`{"direction":"outbound","country":"GH"}`), OccurredAt: occurredAt,
	}
	service := NewService(&fakeStore{createErr: pgx.ErrNoRows, existing: existing})

	result, err := service.Record(t.Context(), organizationID, RecordInput{
		MeterID: meterID, Quantity: 60, SourceType: "voice_call", SourceID: "call-1",
		IdempotencyKey: "voice:call-1", Dimensions: []byte(`{"country":"GH","direction":"outbound"}`),
		OccurredAt: occurredAt,
	})
	if err != nil {
		t.Fatalf("record replay: %v", err)
	}
	if !result.Replayed {
		t.Fatal("matching duplicate must be marked replayed")
	}
}

func TestRecordRejectsConflictingReplay(t *testing.T) {
	organizationID := uuid.New()
	meterID := uuid.New()
	occurredAt := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	existing := UsageEvent{
		ID: uuid.New(), OrganizationID: organizationID, MeterID: meterID, Quantity: 60,
		SourceType: "voice_call", SourceID: "call-1", IdempotencyKey: "voice:call-1",
		Dimensions: []byte(`{}`), OccurredAt: occurredAt,
	}
	service := NewService(&fakeStore{createErr: pgx.ErrNoRows, existing: existing})

	_, err := service.Record(t.Context(), organizationID, RecordInput{
		MeterID: meterID, Quantity: 120, SourceType: "voice_call", SourceID: "call-1",
		IdempotencyKey: "voice:call-1", OccurredAt: occurredAt,
	})
	if !errors.Is(err, ErrUsageEventConflict) {
		t.Fatalf("expected usage conflict, got %v", err)
	}
}

func TestRecordRejectsInvalidUsage(t *testing.T) {
	service := NewService(&fakeStore{})

	_, err := service.Record(t.Context(), uuid.New(), RecordInput{})
	if !errors.Is(err, ErrInvalidUsageEvent) {
		t.Fatalf("expected invalid usage event, got %v", err)
	}
}
