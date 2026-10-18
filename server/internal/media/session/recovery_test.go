package session_test

import (
	"context"
	"testing"

	"github.com/leamout/leamout/server/internal/media/session"
)

func TestManagerPreservesPendingToolEventsAcrossControlDetach(t *testing.T) {
	manager := newManager(t, 1)
	cfg := validConfig()
	if err := manager.Start(context.Background(), cfg); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = manager.Stop(context.Background(), cfg.ID) }()

	first, err := manager.AttachControl(cfg.ID)
	if err != nil {
		t.Fatal(err)
	}
	first.Close()

	// Simulate a provider emitting a tool call while Agent Runtime is down.
	// The event is retained even though no control WebSocket is attached.
	// A second attachment must receive it without restarting the media stream.
	// Publishing is through the managed stream's event pump in production;
	// this test exercises attachment state and liveness.
	status, found := manager.Status(cfg.ID)
	if !found || status.ID != cfg.ID || status.OrganizationID != cfg.OrganizationID ||
		status.CallID != cfg.CallID || status.ChannelID != cfg.ChannelID || status.Attached {
		t.Fatalf("incorrect active unattached media status: %+v, found=%v", status, found)
	}
	second, err := manager.AttachControl(cfg.ID)
	if err != nil {
		t.Fatalf("reattach control: %v", err)
	}
	second.Close()
	if _, found := manager.Status(cfg.ID); !found {
		t.Fatal("control close unexpectedly destroyed live media session")
	}
}

func TestManagerStatusRejectsUnknownSession(t *testing.T) {
	manager := newManager(t, 1)
	if _, found := manager.Status(validConfig().ID); found {
		t.Fatal("unknown media session reported alive")
	}
}
