package wallets

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

func TestSameReservationRequestUsesPostgreSQLTimestampPrecision(t *testing.T) {
	expiresAt := time.Date(2026, time.September, 11, 12, 30, 0, 123456789, time.UTC)
	input := ReserveInput{
		AmountMinor:   2500,
		OperationType: "managed_call",
		OperationID:   "call-123",
		ExpiresAt:     expiresAt,
	}
	reservation := Reservation{
		AmountMinor:   input.AmountMinor,
		OperationType: input.OperationType,
		OperationID:   input.OperationID,
		ExpiresAt:     expiresAt.Truncate(time.Microsecond),
	}

	if !sameReservationRequest(reservation, input) {
		t.Fatal("sameReservationRequest() = false, want true for a persisted retry")
	}

	reservation.AmountMinor++
	if sameReservationRequest(reservation, input) {
		t.Fatal("sameReservationRequest() = true for conflicting amount")
	}
}

func TestCompletedReservationTransitionsAreIdempotentOnlyWhenEquivalent(t *testing.T) {
	capturedAmount := int64(1800)
	captured := Reservation{Status: ReservationCaptured, CapturedAmountMinor: &capturedAmount}
	if !sameCapture(captured, capturedAmount) {
		t.Fatal("sameCapture() = false for equivalent retry")
	}
	if sameCapture(captured, capturedAmount+1) {
		t.Fatal("sameCapture() = true for conflicting amount")
	}
	if sameRelease(captured) {
		t.Fatal("sameRelease() = true for captured reservation")
	}
	if !sameRelease(Reservation{Status: ReservationReleased}) {
		t.Fatal("sameRelease() = false for released reservation")
	}
}
