package freeswitch

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// A bgapi originate may reach FreeSWITCH even when the TCP connection closes
// before Leamout receives its acknowledgement. Preserve the caller-supplied
// UUID and return an uncertain outcome; never issue a second originate.
func TestOriginateLostAcknowledgementPreservesChannelIdentity(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = listener.Close() }()

	commandReceived := make(chan string, 1)
	serverErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer func() { _ = conn.Close() }()

		if _, err := io.WriteString(conn, "Content-Type: auth/request\n\n"); err != nil {
			serverErr <- err
			return
		}
		reader := bufio.NewReader(conn)
		auth, err := reader.ReadString('\n')
		if err != nil {
			serverErr <- err
			return
		}
		if !strings.HasPrefix(auth, "auth ") {
			serverErr <- fmt.Errorf("expected auth, got %q", auth)
			return
		}
		if _, err = reader.ReadString('\n'); err != nil {
			serverErr <- err
			return
		}
		if _, err = io.WriteString(conn, "Content-Type: command/reply\nReply-Text: +OK accepted\n\n"); err != nil {
			serverErr <- err
			return
		}

		command, err := reader.ReadString('\n')
		if err != nil {
			serverErr <- err
			return
		}
		// We deliberately lose the bgapi response after accepting its command.
		commandReceived <- command
		serverErr <- nil
	}()

	cfg := DefaultConfig(listener.Addr().String(), "test-password")
	cfg.ConnectTimeout = time.Second
	cfg.CommandTimeout = time.Second
	client, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = client.Close() }()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Connect(ctx); err != nil {
		t.Fatal(err)
	}

	channelID := uuid.NewString()
	call, originateErr := client.Originate(ctx, OriginateRequest{
		ChannelID:   channelID,
		Endpoint:    "sofia/internal/+14155550100@opensips:5060",
		Destination: "+14155550100",
	})
	if !errors.Is(originateErr, ErrOriginateUncertain) {
		t.Fatalf("expected uncertain bgapi outcome, got %v", originateErr)
	}
	if call.UUID != channelID {
		t.Fatalf("lost channel identity: got %q, want %q", call.UUID, channelID)
	}
	select {
	case command := <-commandReceived:
		if !strings.HasPrefix(command, "bgapi originate ") {
			t.Fatalf("expected one bgapi originate, got %q", command)
		}
		if !strings.Contains(command, "origination_uuid="+channelID) {
			t.Fatalf("bgapi did not carry immutable channel UUID: %q", command)
		}
	case <-ctx.Done():
		t.Fatal("originate command was not sent")
	}
	select {
	case err := <-serverErr:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("fake ESL server did not complete")
	}
}
