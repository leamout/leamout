package prepaid

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPostRequiresReservationForCapture(t *testing.T) {
	service := NewService(nil)
	_, err := service.Post(context.Background(), uuid.New(), uuid.New(), PostEntryInput{Type: EntryCapture})
	if !errors.Is(err, ErrReservationRequired) {
		t.Fatalf("Post() error = %v, want %v", err, ErrReservationRequired)
	}
}

func TestReserveRejectsInvalidMoneyBeforeDatabaseAccess(t *testing.T) {
	service := NewService(nil)
	tests := []ReserveInput{
		{AmountMinor: 0, ExpiresAt: time.Now().Add(time.Hour)},
		{AmountMinor: 100, ExpiresAt: time.Now().Add(-time.Hour)},
	}
	for _, input := range tests {
		_, err := service.Reserve(context.Background(), uuid.New(), uuid.New(), input)
		if !errors.Is(err, ErrInvalidMoney) {
			t.Fatalf("Reserve() error = %v, want %v", err, ErrInvalidMoney)
		}
	}
}

func TestIncreaseRejectsInvalidMoneyBeforeDatabaseAccess(t *testing.T) {
	service := NewService(nil)
	tests := []IncreaseReservationInput{
		{AmountMinor: 0, ExpiresAt: time.Now().Add(time.Hour)},
		{AmountMinor: 100, ExpiresAt: time.Now().Add(-time.Hour)},
	}
	for _, input := range tests {
		_, err := service.Increase(context.Background(), uuid.New(), uuid.New(), input)
		if !errors.Is(err, ErrInvalidMoney) {
			t.Fatalf("Increase() error = %v, want %v", err, ErrInvalidMoney)
		}
	}
}
