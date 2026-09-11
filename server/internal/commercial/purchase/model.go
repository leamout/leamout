package purchase

import (
	"time"

	"github.com/google/uuid"
)

type CheckoutStatus string
type CheckoutType string

const (
	CheckoutSucceeded   CheckoutStatus = "succeeded"
	CheckoutFailed      CheckoutStatus = "failed"
	CheckoutCancelled   CheckoutStatus = "cancelled"
	CheckoutWalletTopup CheckoutType   = "wallet_topup"
)

type FulfillInput struct {
	OrganizationID        uuid.UUID
	CheckoutID            uuid.UUID
	PaymentID             uuid.UUID
	WalletID              *uuid.UUID
	Type                  CheckoutType
	ExpectedCheckoutState string
	AmountMinor           int64
	CompletedAt           time.Time
}

type Result struct {
	OrderID uuid.UUID
	Applied bool
}
