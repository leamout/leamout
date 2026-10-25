package freeswitch

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestEventDispatchPreservesPerChannelOrder(t *testing.T) {
	client, err := New(DefaultConfig("127.0.0.1:8021", "test"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()

	const count = 25
	received := make(chan string, count)
	client.subscriptions = []subscription{{
		events: []string{"CHANNEL_CREATE"}, handler: func(_ context.Context, event Event) error {
			received <- event.Header("Event-Sequence")
			return nil
		},
	}}

	for i := 0; i < count; i++ {
		frame := Frame{ContentType: ContentTypeEventPlain, Headers: map[string]string{
			"Event-Name": "CHANNEL_CREATE", "Unique-ID": "same-channel",
			"Event-Sequence": string(rune('A' + i)),
		}}
		if err := client.dispatchEvent(frame); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < count; i++ {
		select {
		case seq := <-received:
			if seq != string(rune('A'+i)) {
				t.Fatalf("event order %d: %q", i, seq)
			}
		case <-time.After(3 * time.Second):
			t.Fatal("timed out waiting for ordered events")
		}
	}
}

func TestEventDispatchOverloadIsExplicitAndBounded(t *testing.T) {
	client, err := New(DefaultConfig("127.0.0.1:8021", "test"))
	if err != nil {
		t.Fatal(err)
	}
	release := make(chan struct{})
	entered := make(chan struct{})
	var once sync.Once
	client.subscriptions = []subscription{{
		events: []string{"CHANNEL_ANSWER"}, handler: func(ctx context.Context, _ Event) error {
			once.Do(func() { close(entered) })
			select {
			case <-release:
			case <-ctx.Done():
			}
			return nil
		},
	}}
	frame := Frame{ContentType: ContentTypeEventPlain, Headers: map[string]string{"Event-Name": "CHANNEL_ANSWER", "Unique-ID": "blocked-channel"}}
	if err := client.dispatchEvent(frame); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("dispatch worker did not start")
	}
	for i := 0; i < eventQueueDepth; i++ {
		if err := client.dispatchEvent(frame); err != nil {
			t.Fatalf("queue filled early at %d: %v", i, err)
		}
	}
	if err := client.dispatchEvent(frame); err == nil || !strings.Contains(err.Error(), "queue full") {
		t.Fatalf("expected overload instead of silent event loss; got %v", err)
	}
	close(release)
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestEventDispatchStableShard(t *testing.T) {
	for _, key := range []string{"abc", "another-channel", "unkeyed"} {
		want := eventDispatchShard(key)
		if want < 0 || want >= eventDispatchShards {
			t.Fatalf("invalid shard %d for %q", want, key)
		}
		for i := 0; i < 10; i++ {
			if got := eventDispatchShard(key); got != want {
				t.Fatalf("unstable shard for %q: got %d, want %d", key, got, want)
			}
		}
	}
}
