package coturn

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestNewRejectsInvalidConfig(t *testing.T) {
	tests := []Config{
		{},
		{Address: "coturn", Timeout: time.Second},
		{Address: "coturn:3478"},
	}

	for _, config := range tests {
		if _, err := New(config); err == nil {
			t.Fatalf("New(%+v) succeeded, want error", config)
		}
	}
}

func TestHealthCheck(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	client, err := New(Config{Address: listener.Addr().String(), Timeout: time.Second})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := client.HealthCheck(ctx); err != nil {
		t.Fatalf("HealthCheck: %v", err)
	}
}
