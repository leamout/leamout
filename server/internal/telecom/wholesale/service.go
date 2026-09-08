package wholesale

import (
	"context"
	"errors"
	"regexp"
	"strings"
)

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

type store interface {
	Reconcile(context.Context, CDR) (Result, error)
}

type Service struct{ store store }

func NewService(store store) *Service { return &Service{store: store} }

func (s *Service) Reconcile(ctx context.Context, cdr CDR) (Result, error) {
	cdr.Provider = strings.ToLower(strings.TrimSpace(cdr.Provider))
	cdr.ProviderRecordID = strings.TrimSpace(cdr.ProviderRecordID)
	cdr.Direction = strings.ToLower(strings.TrimSpace(cdr.Direction))
	cdr.SIPCallID = strings.TrimSpace(cdr.SIPCallID)
	cdr.Currency = strings.ToUpper(strings.TrimSpace(cdr.Currency))
	if cdr.Provider == "" || cdr.ProviderRecordID == "" || cdr.Direction != "termination" || cdr.SIPCallID == "" ||
		cdr.StartedAt.IsZero() || cdr.DurationSeconds < 0 || cdr.CostMicros < 0 ||
		!currencyPattern.MatchString(cdr.Currency) || cdr.Raw == nil {
		return Result{}, ErrInvalidCDR
	}
	if s == nil || s.store == nil {
		return Result{}, errors.New("provider CDR reconciliation unavailable")
	}
	return s.store.Reconcile(ctx, cdr)
}
