package payments

import (
	"context"
	"encoding/json"
)

type eventStore interface {
	processProviderEvent(context.Context, ProviderEvent) (Settlement, error)
	checkProviderEventReplay(context.Context, ProviderEvent) error
}

// Service owns normalized payment-provider event validation and reconciliation.
// Provider adapters authenticate and parse payloads before they reach this boundary.
type Service struct {
	events eventStore
}

func NewService(events eventStore) *Service { return &Service{events: events} }

func (s *Service) ProcessProviderEvent(ctx context.Context, event ProviderEvent) (Settlement, error) {
	relevant, valid := classifyProviderEvent(event)
	if !valid {
		return Settlement{}, ErrPaymentMismatch
	}
	if !relevant {
		return Settlement{}, s.events.checkProviderEventReplay(ctx, event)
	}
	return s.events.processProviderEvent(ctx, event)
}

func classifyProviderEvent(event ProviderEvent) (relevant, valid bool) {
	if event.Provider == "" || event.ProviderEventID == "" || len(event.Raw) == 0 || !json.Valid(event.Raw) {
		return false, false
	}
	switch event.Provider {
	case "stripe":
		relevant = event.Type == "checkout.session.completed" || event.Type == "checkout.session.expired"
	case "paystack":
		relevant = event.Type == "charge.success" || event.Type == "charge.failed"
	default:
		return false, false
	}
	if !relevant {
		return false, true
	}
	if event.Payment.Reference == "" || !eventTypeMatchesStatus(event) {
		return true, false
	}
	return true, true
}

func eventTypeMatchesStatus(event ProviderEvent) bool {
	switch event.Provider + ":" + event.Type {
	case "stripe:checkout.session.completed", "paystack:charge.success":
		return event.Payment.Status == StatusSucceeded
	case "stripe:checkout.session.expired":
		return event.Payment.Status == StatusCancelled
	case "paystack:charge.failed":
		return event.Payment.Status == StatusFailed || event.Payment.Status == StatusCancelled
	default:
		return false
	}
}
