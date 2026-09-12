package prepaid

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

type Status string

const (
	StatusActive     Status = "active"
	StatusRestricted Status = "restricted"
	StatusClosed     Status = "closed"
)

type EntryType string

const (
	EntryTopup            EntryType = "topup"
	EntryCapture          EntryType = "capture"
	EntryRefund           EntryType = "refund"
	EntryChargeback       EntryType = "chargeback"
	EntryAdjustmentCredit EntryType = "adjustment_credit"
	EntryAdjustmentDebit  EntryType = "adjustment_debit"
)

type ReservationStatus string

const (
	ReservationActive   ReservationStatus = "active"
	ReservationCaptured ReservationStatus = "captured"
	ReservationReleased ReservationStatus = "released"
	ReservationExpired  ReservationStatus = "expired"
)

var (
	ErrWalletNotFound          = apperror.NewNotFound("wallet not found")
	ErrReservationNotFound     = apperror.NewNotFound("wallet reservation not found")
	ErrWalletExists            = apperror.NewConflict("wallet already exists for currency")
	ErrInsufficientFunds       = apperror.NewConflict("insufficient prepaid funds")
	ErrInvalidReservationState = apperror.NewConflict("wallet reservation is no longer active")
	ErrDuplicateOperation      = apperror.NewConflict("wallet operation already reserved")
	ErrDuplicateLedgerEntry    = apperror.NewConflict("wallet ledger entry already exists")
	ErrReservationRequired     = apperror.NewConflict("wallet capture requires an active reservation")
	ErrInvalidMoney            = apperror.NewBadRequest("invalid monetary amount or currency")
)

type Wallet struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	Currency       string
	Status         Status
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Balance struct {
	PostedMinor    int64
	ReservedMinor  int64
	AvailableMinor int64
}

type LedgerEntry struct {
	ID             uuid.UUID
	WalletID       uuid.UUID
	OrganizationID uuid.UUID
	Type           EntryType
	AmountMinor    int64
	SourceType     string
	SourceID       string
	IdempotencyKey string
	Metadata       json.RawMessage
	OccurredAt     time.Time
	CreatedAt      time.Time
}

type Reservation struct {
	ID                  uuid.UUID
	WalletID            uuid.UUID
	OrganizationID      uuid.UUID
	AmountMinor         int64
	CapturedAmountMinor *int64
	OperationType       string
	OperationID         string
	Status              ReservationStatus
	ExpiresAt           time.Time
	CapturedAt          *time.Time
	ReleasedAt          *time.Time
	ExpiredAt           *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type PostEntryInput struct {
	Type           EntryType
	AmountMinor    int64
	SourceType     string
	SourceID       string
	IdempotencyKey string
	Metadata       json.RawMessage
	OccurredAt     *time.Time
}

type ReserveInput struct {
	AmountMinor   int64
	OperationType string
	OperationID   string
	ExpiresAt     time.Time
}

type IncreaseReservationInput struct {
	AmountMinor int64
	ExpiresAt   time.Time
}
