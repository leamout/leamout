package orders

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

type Type string

const (
	TypeSubscription Type = "subscription"
	TypeWalletTopup  Type = "wallet_topup"
)

var ErrOrderNotFound = apperror.NewNotFound("order not found")

type Order struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	CheckoutID     uuid.UUID
	PaymentID      uuid.UUID
	WalletID       *uuid.UUID
	PriceID        *uuid.UUID
	Type           Type
	AmountMinor    int64
	Currency       string
	CompletedAt    time.Time
	Metadata       json.RawMessage
	CreatedAt      time.Time
}
