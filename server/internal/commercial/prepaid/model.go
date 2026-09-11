package prepaid

import (
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/pkg/apperror"
)

var ErrInvalidManagedOperation = apperror.NewBadRequest("invalid managed operation authorization")

type Status string

const (
	StatusAuthorized Status = "authorized"
	StatusCaptured   Status = "captured"
	StatusReleased   Status = "released"
	StatusExpired    Status = "expired"
)

// AuthorizeInput contains the customer-facing monetary terms that Commercial
// has resolved before a managed provider operation may begin.
type AuthorizeInput struct {
	OrganizationID uuid.UUID
	Currency       string
	AmountMinor    int64
	OperationType  string
	OperationID    string
	ExpiresAt      time.Time
}

// CaptureInput settles an authorized managed operation for its final amount.
type CaptureInput struct {
	OrganizationID  uuid.UUID
	AuthorizationID uuid.UUID
	AmountMinor     int64
}

// Authorization is the provider-neutral proof that prepaid funds were held.
// Provider adapters need only its ID; wallet identity remains Commercial state.
type Authorization struct {
	ID             uuid.UUID
	OrganizationID uuid.UUID
	AmountMinor    int64
	CapturedMinor  *int64
	OperationType  string
	OperationID    string
	Status         Status
	ExpiresAt      time.Time
}
