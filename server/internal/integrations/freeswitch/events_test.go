package freeswitch

import "testing"

func TestPlainEventHeaders(t *testing.T) {
	frame := Frame{
		ContentType: ContentTypeEventPlain,
		Headers: map[string]string{
			"Content-Type": ContentTypeEventPlain,
		},
		Body: "Event-Name: CHANNEL_CREATE\nUnique-ID: channel-123\nCaller-Caller-ID-Number: %2B233201234567\n",
	}

	if got := frame.Header("Event-Name"); got != "CHANNEL_CREATE" {
		t.Fatalf("Event-Name = %q, want CHANNEL_CREATE", got)
	}

	event := Event{
		Name:    frame.Header("Event-Name"),
		Headers: frame.Headers,
		Body:    frame.Body,
	}
	if got := event.Header("Unique-ID"); got != "channel-123" {
		t.Fatalf("Unique-ID = %q, want channel-123", got)
	}
	if got := event.Header("Caller-Caller-ID-Number"); got != "+233201234567" {
		t.Fatalf("caller = %q, want +233201234567", got)
	}
}
