package billing

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	polarintegration "github.com/leamout/leamout/server/internal/integrations/payments/polar"
	"github.com/leamout/leamout/server/pkg/apperror"
)

// Consumer decodes verified Polar events and dispatches them to billing use cases.
type Consumer struct {
	service *Service
}

func NewConsumer(service *Service) *Consumer {
	return &Consumer{
		service: service,
	}
}

func (c *Consumer) Handle(ctx context.Context, payload []byte) (bool, error) {
	var event polarintegration.Event
	if err := json.Unmarshal(payload, &event); err != nil {
		return false, apperror.NewBadRequest("invalid stored Polar webhook")
	}

	switch event.Type {
	case "subscription.created", "subscription.updated", "subscription.active",
		"subscription.canceled", "subscription.uncanceled", "subscription.cycled",
		"subscription.revoked", "subscription.past_due", "subscription.paused",
		"subscription.resumed", "subscription.migrated":
		var value polarintegration.Subscription
		if err := json.Unmarshal(event.Data, &value); err != nil {
			return false, apperror.NewBadRequest("invalid Polar subscription event")
		}
		id, err := uuid.Parse(value.ProductID)
		if err != nil || id == uuid.Nil {
			return false, apperror.NewBadRequest("Polar subscription product identity is missing")
		}
		if _, err := c.service.planCodeFromProduct(value); err != nil {
			return true, nil
		}
		return false, c.service.syncPolarSubscription(ctx, event.Timestamp, value)
	case "customer.state_changed":
		var state polarintegration.CustomerState
		if err := json.Unmarshal(event.Data, &state); err != nil {
			return false, apperror.NewBadRequest("invalid Polar customer state event")
		}
		id, err := uuid.Parse(state.ExternalID)
		if err != nil || id == uuid.Nil {
			return true, nil
		}
		return false, c.service.RecoverOrganization(ctx, id)
	default:
		return true, nil
	}
}
