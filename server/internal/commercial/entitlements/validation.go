package entitlements

import (
	"errors"
	"strings"
)

var (
	ErrKeyRequired   = errors.New("entitlement key is required")
	ErrInvalidKey    = errors.New("entitlement key must not contain whitespace")
	ErrInvalidKind   = errors.New("entitlement kind is invalid")
	ErrInvalidValue  = errors.New("entitlement value does not match its kind")
	ErrInvalidPeriod = errors.New("entitlement expiration must not precede its start")
)

// Validate checks an entitlement before it is persisted or signed.
func Validate(entitlement Entitlement) error {
	if strings.TrimSpace(entitlement.Key) == "" {
		return ErrKeyRequired
	}
	if strings.IndexFunc(entitlement.Key, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }) >= 0 {
		return ErrInvalidKey
	}
	switch entitlement.Kind {
	case KindFeature:
		if entitlement.Enabled == nil || entitlement.Limit != nil {
			return ErrInvalidValue
		}
	case KindLimit:
		if entitlement.Enabled != nil || entitlement.Limit == nil || *entitlement.Limit < 0 {
			return ErrInvalidValue
		}
	default:
		return ErrInvalidKind
	}
	if entitlement.StartsAt != nil && entitlement.ExpiresAt != nil && entitlement.ExpiresAt.Before(*entitlement.StartsAt) {
		return ErrInvalidPeriod
	}
	return nil
}
