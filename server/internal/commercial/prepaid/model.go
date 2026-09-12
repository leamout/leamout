package prepaid

import "github.com/leamout/leamout/pkg/apperror"

var (
	ErrManagedNumberPriceUnavailable = apperror.NewServiceUnavailable(
		"managed number purchase price is unavailable",
		nil,
	)
	ErrManagedNumberSubscriptionInactive = apperror.NewConflict(
		"managed number purchase requires an active subscription",
	)
	ErrManagedNumberQuoteExpired = apperror.NewConflict(
		"managed number purchase quote is no longer current",
	)
	ErrManagedNumberAuthorizationInvalid = apperror.NewConflict(
		"managed number purchase authorization is invalid",
	)
)
