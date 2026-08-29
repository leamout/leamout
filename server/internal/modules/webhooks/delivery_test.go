package webhooks

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/leamout/leamout/internal/database/sqlc"
)

func TestHTTPSenderSendsSignedEnvelope(t *testing.T) {
	t.Parallel()

	eventID := uuid.New()
	secret := []byte("01234567890123456789012345678901")
	occurredAt := time.Date(2026, 8, 29, 6, 0, 0, 0, time.UTC)
	var gotEnvelope DeliveryEnvelope
	var gotEventID string
	var gotEventType string
	var gotSignature string

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(body, &gotEnvelope); err != nil {
			t.Errorf("decode envelope: %v", err)
		}
		gotEventID = r.Header.Get("X-Leamout-Event-ID")
		gotEventType = r.Header.Get("X-Leamout-Event")
		gotSignature = r.Header.Get(signatureHeader)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	sender := NewHTTPSender()
	sender.client = server.Client()
	attempt := sender.Send(context.Background(), sqlc.ClaimWebhookDeliveriesRow{
		EventID:       eventID,
		EventType:     "call.answered",
		Payload:       []byte("{\"call_id\":\"call-1\"}"),
		OccurredAt:    pgtype.Timestamptz{Time: occurredAt, Valid: true},
		Url:           server.URL,
		SigningSecret: secret,
	})

	if !deliverySucceeded(attempt) {
		t.Fatalf("delivery attempt = %+v", attempt)
	}
	if gotEnvelope.ID != eventID || gotEnvelope.Type != "call.answered" || !gotEnvelope.OccurredAt.Equal(occurredAt) {
		t.Fatalf("envelope = %+v", gotEnvelope)
	}
	if gotEventID != eventID.String() || gotEventType != "call.answered" {
		t.Fatalf("headers event_id=%q event_type=%q", gotEventID, gotEventType)
	}
	if gotSignature == "" {
		t.Fatal("signature header is empty")
	}
}

func TestDeliveryRetryPolicy(t *testing.T) {
	t.Parallel()

	status := func(value int32) *int32 { return &value }
	cases := []struct {
		name      string
		attempt   DeliveryAttempt
		retryable bool
	}{
		{name: "network", attempt: DeliveryAttempt{Err: context.DeadlineExceeded}, retryable: true},
		{name: "timeout", attempt: DeliveryAttempt{StatusCode: status(408)}, retryable: true},
		{name: "rate limit", attempt: DeliveryAttempt{StatusCode: status(429)}, retryable: true},
		{name: "server", attempt: DeliveryAttempt{StatusCode: status(503)}, retryable: true},
		{name: "bad request", attempt: DeliveryAttempt{StatusCode: status(400)}, retryable: false},
		{name: "not found", attempt: DeliveryAttempt{StatusCode: status(404)}, retryable: false},
	}
	for _, tc := range cases {
		if got := retryableDelivery(tc.attempt); got != tc.retryable {
			t.Fatalf("%s retryable = %v, want %v", tc.name, got, tc.retryable)
		}
	}

	job := &DeliveryJob{config: DeliveryJobConfig{RetryBaseDelay: 2 * time.Second, RetryMaxDelay: 30 * time.Second}}
	if got := job.retryDelay(1); got != 2*time.Second {
		t.Fatalf("retryDelay(1) = %s", got)
	}
	if got := job.retryDelay(5); got != 30*time.Second {
		t.Fatalf("retryDelay(5) = %s", got)
	}
}
