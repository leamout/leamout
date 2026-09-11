package payments

import (
	"context"
	"encoding/json"

	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
)

type eventStore interface {
	processProviderEvent(context.Context, paymentprovider.Event) (Settlement, error)
	checkProviderEventReplay(context.Context, paymentprovider.Event) error
}

// Service owns normalized payment-provider event validation and reconciliation.
// Provider adapters authenticate and parse payloads before they reach this boundary.
type Service struct {
	events eventStore
}

func NewService(events eventStore) *Service { return &Service{events: events} }

func (s *Service) ProcessProviderEvent(ctx context.Context, event paymentprovider.Event) (Settlement, error) {
	relevant, valid := classifyProviderEvent(event)
	if !valid {
		return Settlement{}, ErrPaymentMismatch
	}
	if !relevant {
		return Settlement{}, s.events.checkProviderEventReplay(ctx, event)
	}
	return s.events.processProviderEvent(ctx, event)
}

func classifyProviderEvent(event paymentprovider.Event) (relevant, valid bool) {
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

func eventTypeMatchesStatus(event paymentprovider.Event) bool {
	switch event.Provider + ":" + event.Type {
	case "stripe:checkout.session.completed", "paystack:charge.success":
		return event.Payment.Status == paymentprovider.StatusSucceeded
	case "stripe:checkout.session.expired":
		return event.Payment.Status == paymentprovider.StatusCancelled
	case "paystack:charge.failed":
		return event.Payment.Status == paymentprovider.StatusFailed || event.Payment.Status == paymentprovider.StatusCancelled
	default:
		return false
	}
}
