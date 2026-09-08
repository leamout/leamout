package provider_diagnostics

import (
	"context"
	"testing"
)

type diagnosticsStoreStub struct {
	limit int32
}

func (s *diagnosticsStoreStub) Snapshot(context.Context, int32) (Snapshot, error) {
	return Snapshot{}, nil
}

func TestServiceSnapshotUsesDefaultLimit(t *testing.T) {
	store := &diagnosticsStoreCapture{}
	service := NewService(store)
	if _, err := service.Snapshot(context.Background(), 0); err != nil {
		t.Fatal(err)
	}
	if store.limit != 50 {
		t.Fatalf("limit = %d, want 50", store.limit)
	}
}

func TestServiceSnapshotCapsLimit(t *testing.T) {
	store := &diagnosticsStoreCapture{}
	service := NewService(store)
	if _, err := service.Snapshot(context.Background(), 500); err != nil {
		t.Fatal(err)
	}
	if store.limit != 200 {
		t.Fatalf("limit = %d, want 200", store.limit)
	}
}

type diagnosticsStoreCapture struct {
	limit int32
}

func (s *diagnosticsStoreCapture) Snapshot(_ context.Context, limit int32) (Snapshot, error) {
	s.limit = limit
	return Snapshot{}, nil
}
