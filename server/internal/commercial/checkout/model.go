package checkout

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

type Type string
type Provider string
type PaymentMethod string
type Status string
type NextAction string

const (
	TypeSubscription   Type          = "subscription"
	TypeWalletTopup    Type          = "wallet_topup"
	ProviderStripe     Provider      = "stripe"
	ProviderPaystack   Provider      = "paystack"
	MethodCard         PaymentMethod = "card"
	MethodMobileMoney  PaymentMethod = "mobile_money"
	StatusPending      Status        = "pending"
	StatusProcessing   Status        = "processing"
	StatusSucceeded    Status        = "succeeded"
	StatusFailed       Status        = "failed"
	StatusCancelled    Status        = "cancelled"
	StatusExpired      Status        = "expired"
	ActionNone         NextAction    = "none"
	ActionWait         NextAction    = "wait"
	ActionAuthorizeMoMo NextAction   = "authorize_mobile_money"
	ActionSubmitOTP    NextAction    = "submit_otp"
	ActionSubmitPhone  NextAction    = "submit_phone"
	ActionUnsupported  NextAction    = "unsupported"
)

var (
	ErrCheckoutNotFound = apperror.NewNotFound("checkout not found")
	ErrReferenceConflict = apperror.NewConflict("checkout reference already exists")
	ErrInvalidTransition = apperror.NewConflict("invalid checkout transition")
	ErrInvalidCheckout   = apperror.NewBadRequest("invalid checkout")
)

type Checkout struct {
	ID              uuid.UUID
	OrganizationID  uuid.UUID
	WalletID        *uuid.UUID
	PriceID         *uuid.UUID
	Type            Type
	Provider        Provider
	PaymentMethod   PaymentMethod
	Reference       string
	AmountMinor     int64
	Currency        string
	Status          Status
	NextAction      NextAction
	ProviderMessage *string
	ExpiresAt       time.Time
	CompletedAt     *time.Time
	Metadata        json.RawMessage
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CreateInput struct {
	WalletID      *uuid.UUID
	PriceID       *uuid.UUID
	Type          Type
	Provider      Provider
	PaymentMethod PaymentMethod
	Reference     string
	AmountMinor   int64
	Currency      string
	ExpiresAt     time.Time
	Metadata      json.RawMessage
}

type Transition struct {
	Expected        Status
	Status          Status
	NextAction      NextAction
	ProviderMessage *string
	CompletedAt     *time.Time
}
