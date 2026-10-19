package recordings

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/integrations/freeswitch"
)

func TestRecordingLifecycleEventMapsRecordStart(t *testing.T) {
	timestamp := time.Date(2026, time.August, 29, 8, 0, 0, 0, time.UTC)
	callID := uuid.New()

	event, err := recordingLifecycleEvent(freeswitch.Event{
		Name: freeSWITCHEventRecordStart,
		Headers: map[string]string{
			"Unique-ID":                "channel-1",
			"variable_leamout_call_id": callID.String(),
			"Record-File-Path":         "/var/lib/freeswitch/recordings/call-1.wav",
			"Event-Date-Timestamp":     "1787990400000000",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if event.CallID != callID {
		t.Fatalf("call id = %s, want %s", event.CallID, callID)
	}
	if event.ChannelID != "channel-1" {
		t.Fatalf("channel = %q, want channel-1", event.ChannelID)
	}
	if event.Path != "/var/lib/freeswitch/recordings/call-1.wav" {
		t.Fatalf("path = %q", event.Path)
	}
	if !event.OccurredAt.Equal(timestamp) {
		t.Fatalf("occurred_at = %s, want %s", event.OccurredAt, timestamp)
	}
}

func TestRecordingLifecycleEventMapsRecordStop(t *testing.T) {
	callID := uuid.New()

	event, err := recordingLifecycleEvent(freeswitch.Event{
		Name: freeSWITCHEventRecordStop,
		Headers: map[string]string{
			"Unique-ID":                "channel-2",
			"variable_leamout_call_id": callID.String(),
			"Record-File-Path":         "/var/lib/freeswitch/recordings/call-2.wav",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if event.CallID != callID || event.ChannelID != "channel-2" ||
		event.Path != "/var/lib/freeswitch/recordings/call-2.wav" {
		t.Fatalf("record stop event = %+v", event)
	}
}

func TestConsumerIgnoresUnrelatedEvent(t *testing.T) {
	// Unrelated events must not touch the service dependency.
	consumer := &Consumer{}
	if err := consumer.HandleFreeSWITCHEvent(
		context.Background(),
		freeswitch.Event{Name: "CHANNEL_ANSWER"},
	); err != nil {
		t.Fatal(err)
	}
}

func TestNewConsumerRequiresService(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewConsumer accepted nil service")
		}
	}()
	NewConsumer(nil)
}

func TestRecordingLifecycleEventRequiresIdentity(t *testing.T) {
	callID := uuid.New().String()
	tests := []struct {
		name    string
		headers map[string]string
	}{
		{name: "missing call", headers: map[string]string{
			"Unique-ID":        "channel-3",
			"Record-File-Path": "/tmp/test.wav",
		}},
		{name: "missing channel", headers: map[string]string{
			"variable_leamout_call_id": callID,
			"Record-File-Path":         "/tmp/test.wav",
		}},
		{name: "missing path", headers: map[string]string{
			"variable_leamout_call_id": callID,
			"Unique-ID":                "channel-3",
		}},
		{name: "invalid timestamp", headers: map[string]string{
			"variable_leamout_call_id": callID,
			"Unique-ID":                "channel-3",
			"Record-File-Path":         "/tmp/test.wav",
			"Event-Date-Timestamp":     "not-a-timestamp",
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := recordingLifecycleEvent(freeswitch.Event{
				Name:    freeSWITCHEventRecordStart,
				Headers: tt.headers,
			})
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
