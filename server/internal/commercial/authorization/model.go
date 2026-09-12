package authorization

import "github.com/leamout/leamout/pkg/apperror"

const ManagedVoiceFeature = "voice.managed.enabled"

var (
	ErrManagedNumberPriceUnavailable = apperror.NewServiceUnavailable(
		"managed number purchase price is unavailable",
		nil,
	)
	ErrManagedNumberAccessRequired = apperror.NewConflict(
		"managed number purchase requires active managed-service access",
	)
	ErrManagedNumberQuoteExpired = apperror.NewConflict(
		"managed number purchase quote is no longer current",
	)
	ErrManagedNumberAuthorizationInvalid = apperror.NewConflict(
		"managed number purchase authorization is invalid",
	)
)
