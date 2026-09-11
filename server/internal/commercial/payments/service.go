package payments

import (
	"context"
	"encoding/json"

	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
)

type eventStore interface {
	processProviderEvent(context.Context, paymentprovider.Event) (Settlement, error)
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
		return Settlement{}, nil
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
	return true, event.Payment.Reference != ""
}
