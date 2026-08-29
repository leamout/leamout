package outbox

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/database/sqlc"
)

type fakeMessagePublisher struct {
	subject   string
	payload   []byte
	headers   map[string]string
	messageID string
	err       error
}

func (f *fakeMessagePublisher) PublishWithOptions(_ context.Context, subject string, payload []byte, headers map[string]string, messageID string) error {
	f.subject = subject
	f.payload = append([]byte(nil), payload...)
	f.headers = headers
	f.messageID = messageID
	return f.err
}

func TestPublisherPublishMapsOutboxEventToNATS(t *testing.T) {
	t.Parallel()

	eventID := uuid.New()
	aggregateID := uuid.New()
	client := &fakeMessagePublisher{}
	publisher := NewPublisher(client)
	event := sqlc.OutboxEvent{
		ID:            eventID,
		Subject:       "call.answered",
		AggregateType: "call",
		AggregateID:   aggregateID,
		Payload:       []byte(`{"call_id":"` + aggregateID.String() + `"}`),
		Headers:       []byte(`{"organization_id":"org-1"}`),
	}

	if err := publisher.Publish(context.Background(), event); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	if client.subject != "leamout.event.call.answered" {
		t.Fatalf("subject = %q", client.subject)
	}
	if client.messageID != eventID.String() {
		t.Fatalf("message id = %q", client.messageID)
	}
	if client.headers[headerEventID] != eventID.String() {
		t.Fatalf("event header = %q", client.headers[headerEventID])
	}
	if client.headers[headerAggregateType] != "call" {
		t.Fatalf("aggregate type header = %q", client.headers[headerAggregateType])
	}
	if client.headers[headerAggregateID] != aggregateID.String() {
		t.Fatalf("aggregate id header = %q", client.headers[headerAggregateID])
	}
	if client.headers["organization_id"] != "org-1" {
		t.Fatalf("stored header = %q", client.headers["organization_id"])
	}
}

func TestPublisherPublishReturnsTransportError(t *testing.T) {
	t.Parallel()

	want := errors.New("nats unavailable")
	client := &fakeMessagePublisher{err: want}
	publisher := NewPublisher(client)
	event := sqlc.OutboxEvent{
		ID:            uuid.New(),
		Subject:       "call.failed",
		AggregateType: "call",
		AggregateID:   uuid.New(),
		Payload:       []byte(`{}`),
		Headers:       []byte(`{}`),
	}

	if err := publisher.Publish(context.Background(), event); !errors.Is(err, want) {
		t.Fatalf("Publish() error = %v, want wrapped %v", err, want)
	}
}

func TestPublisherJobRetryDelayIsCapped(t *testing.T) {
	t.Parallel()

	job := &PublisherJob{config: PublisherJobConfig{
		RetryBaseDelay: 2 * time.Second,
		RetryMaxDelay:  30 * time.Second,
	}}

	cases := []struct {
		attempt int32
		want    time.Duration
	}{
		{attempt: 1, want: 2 * time.Second},
		{attempt: 2, want: 4 * time.Second},
		{attempt: 3, want: 8 * time.Second},
		{attempt: 4, want: 16 * time.Second},
		{attempt: 5, want: 30 * time.Second},
		{attempt: 20, want: 30 * time.Second},
	}

	for _, tc := range cases {
		if got := job.retryDelay(tc.attempt); got != tc.want {
			t.Fatalf("retryDelay(%d) = %s, want %s", tc.attempt, got, tc.want)
		}
	}
}
