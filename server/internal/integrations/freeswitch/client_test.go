package freeswitch

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestReadFrameAcceptsBoundedEvent(t *testing.T) {
	body := strings.Repeat("x", maxESLEventBodyBytes)
	wire := fmt.Sprintf("Content-Type: %s\nContent-Length: %d\n\n%s", ContentTypeEventPlain, len(body), body)
	frame, err := readFrame(bufio.NewReader(strings.NewReader(wire)))
	if err != nil {
		t.Fatal(err)
	}
	if frame.Body != body {
		t.Fatalf("body length = %d, want %d", len(frame.Body), len(body))
	}
}

func TestReadFrameRejectsOversizedEventWithoutReadingBody(t *testing.T) {
	wire := fmt.Sprintf("Content-Type: %s\nContent-Length: %d\n\n", ContentTypeEventPlain, maxESLEventBodyBytes+1)
	_, err := readFrame(bufio.NewReader(strings.NewReader(wire)))
	if !errors.Is(err, errESLFrameTooLarge) {
		t.Fatalf("expected event size error, got %v", err)
	}
}

func TestReadFrameAllowsLargerControlReply(t *testing.T) {
	body := strings.Repeat("r", maxESLEventBodyBytes+1)
	wire := fmt.Sprintf("Content-Type: %s\nContent-Length: %d\n\n%s", ContentTypeAPIResponse, len(body), body)
	frame, err := readFrame(bufio.NewReader(strings.NewReader(wire)))
	if err != nil {
		t.Fatal(err)
	}
	if len(frame.Body) != len(body) {
		t.Fatalf("body length = %d, want %d", len(frame.Body), len(body))
	}
}

func TestReadFrameRejectsOversizedControlReply(t *testing.T) {
	wire := fmt.Sprintf("Content-Type: %s\nContent-Length: %d\n\n", ContentTypeAPIResponse, maxESLReplyBodyBytes+1)
	_, err := readFrame(bufio.NewReader(strings.NewReader(wire)))
	if !errors.Is(err, errESLFrameTooLarge) {
		t.Fatalf("expected reply size error, got %v", err)
	}
}

func TestReadFrameRejectsOversizedHeader(t *testing.T) {
	wire := "X-Test: " + strings.Repeat("x", maxESLHeaderBytes) + "\n\n"
	_, err := readFrame(bufio.NewReader(strings.NewReader(wire)))
	if !errors.Is(err, errESLFrameTooLarge) {
		t.Fatalf("expected header size error, got %v", err)
	}
}

func TestReadFrameRejectsOversizedTotalHeaders(t *testing.T) {
	wire := strings.Repeat("X-Key: value\n", maxESLHeaderBytes/10) + "\n"
	_, err := readFrame(bufio.NewReader(strings.NewReader(wire)))
	if !errors.Is(err, errESLFrameTooLarge) {
		t.Fatalf("expected total header size error, got %v", err)
	}
}
