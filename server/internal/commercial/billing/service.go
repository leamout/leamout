package billing

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/commercial/plans"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
	stripeintegration "github.com/leamout/leamout/server/internal/integrations/payments/stripe"
	"github.com/leamout/leamout/server/pkg/apperror"
)

const providerStripe = "stripe"

type Service struct {
	plans         *plans.Service
	subscriptions *subscriptions.Service
	stripe        *stripeintegration.Client
	priceIDs      map[string]string
}

func NewService(
	plansService *plans.Service,
	subscriptionsService *subscriptions.Service,
	stripeClient *stripeintegration.Client,
	priceIDs map[string]string,
) *Service {
	return &Service{
		plans:         plansService,
		subscriptions: subscriptionsService,
		stripe:        stripeClient,
		priceIDs:      priceIDs,
	}
}

func (s *Service) Checkout(
	ctx context.Context,
	organizationID uuid.UUID,
	req CheckoutRequest,
) (SessionResponse, error) {
	if organizationID == uuid.Nil {
		return SessionResponse{}, apperror.NewBadRequest("organization context required")
	}
	if s.stripe == nil || !s.stripe.Configured() {
		return SessionResponse{}, apperror.NewServiceUnavailable("Stripe billing is not configured", nil)
	}

	plan, err := s.plans.GetByCode(ctx, strings.TrimSpace(req.PlanCode))
	if err != nil {
		return SessionResponse{}, err
	}
	if plan.PricingType != plans.PricingTypeFixed || plan.AmountMinor == nil || *plan.AmountMinor == 0 {
		return SessionResponse{}, apperror.NewBadRequest("plan is not available for Stripe checkout")
	}

	priceID := strings.TrimSpace(s.priceIDs[plan.Code])
	if priceID == "" {
		return SessionResponse{}, apperror.NewServiceUnavailable("Stripe price is not configured for this plan", nil)
	}

	existing, err := s.subscriptions.Get(ctx, organizationID)
	if err != nil {
		return SessionResponse{}, err
	}

	customerID := ""
	if existing != nil {
		if existing.Provider != nil &&
			*existing.Provider == providerStripe &&
			existing.ProviderCustomerID != nil {
			customerID = *existing.ProviderCustomerID
		}
		if existing.ProviderSubscriptionID != nil &&
			existing.Status != subscriptions.StatusCanceled {
			return SessionResponse{}, apperror.NewConflict("organization already has an active billing subscription")
		}
	}

	session, err := s.stripe.CreateCheckoutSession(ctx, stripeintegration.CheckoutParams{
		PriceID:        priceID,
		CustomerID:     customerID,
		OrganizationID: organizationID.String(),
		PlanCode:       plan.Code,
		SuccessURL:     req.SuccessURL,
		CancelURL:      req.CancelURL,
	})
	if err != nil {
		return SessionResponse{}, apperror.NewServiceUnavailable("create Stripe checkout session", err)
	}
	return SessionResponse{URL: session.URL}, nil
}

func (s *Service) Portal(
	ctx context.Context,
	organizationID uuid.UUID,
	req PortalRequest,
) (SessionResponse, error) {
	if organizationID == uuid.Nil {
		return SessionResponse{}, apperror.NewBadRequest("organization context required")
	}
	if s.stripe == nil || !s.stripe.Configured() {
		return SessionResponse{}, apperror.NewServiceUnavailable("Stripe billing is not configured", nil)
	}

	existing, err := s.subscriptions.Get(ctx, organizationID)
	if err != nil {
		return SessionResponse{}, err
	}
	if existing == nil ||
		existing.Provider == nil ||
		*existing.Provider != providerStripe ||
		existing.ProviderCustomerID == nil {
		return SessionResponse{}, apperror.NewBadRequest("Stripe billing portal is unavailable for this organization")
	}

	session, err := s.stripe.CreatePortalSession(ctx, stripeintegration.PortalParams{
		CustomerID: *existing.ProviderCustomerID,
		ReturnURL:  req.ReturnURL,
	})
	if err != nil {
		return SessionResponse{}, apperror.NewServiceUnavailable("create Stripe billing portal session", err)
	}
	return SessionResponse{URL: session.URL}, nil
}

func (s *Service) Webhook(ctx context.Context, payload []byte, signature string) error {
	if s.stripe == nil || !s.stripe.Configured() {
		return apperror.NewServiceUnavailable("Stripe billing is not configured", nil)
	}

	event, err := s.stripe.ParseWebhook(payload, signature)
	if err != nil {
		return apperror.NewBadRequest("invalid Stripe webhook")
	}

	switch event.Type {
	case "checkout.session.completed":
		session, decodeErr := stripeintegration.DecodeEventObject[stripeintegration.CheckoutSession](event)
		if decodeErr != nil {
			return apperror.NewBadRequest("invalid Stripe checkout event")
		}
		if session.Subscription == "" {
			return nil
		}
		subscription, getErr := s.stripe.GetSubscription(ctx, session.Subscription)
		if getErr != nil {
			return apperror.NewServiceUnavailable("retrieve Stripe subscription", getErr)
		}
		if len(subscription.Metadata) == 0 {
			subscription.Metadata = session.Metadata
		}
		return s.syncStripeSubscription(ctx, event.Created, subscription)

	case "customer.subscription.created",
		"customer.subscription.updated",
		"customer.subscription.deleted":
		subscription, decodeErr := stripeintegration.DecodeEventObject[stripeintegration.Subscription](event)
		if decodeErr != nil {
			return apperror.NewBadRequest("invalid Stripe subscription event")
		}
		return s.syncStripeSubscription(ctx, event.Created, subscription)
	default:
		return nil
	}
}

func (s *Service) syncStripeSubscription(
	ctx context.Context,
	eventCreated int64,
	value stripeintegration.Subscription,
) error {
	organizationID, err := uuid.Parse(strings.TrimSpace(value.Metadata["organization_id"]))
	if err != nil || organizationID == uuid.Nil {
		return apperror.NewBadRequest("Stripe subscription is missing organization metadata")
	}

	plan, err := s.plans.GetByCode(ctx, strings.TrimSpace(value.Metadata["plan_code"]))
	if err != nil {
		return err
	}

	periodStart, periodEnd := value.CurrentPeriod()
	_, err = s.subscriptions.Sync(ctx, organizationID, subscriptions.SyncRequest{
		PlanID:                 plan.ID,
		Status:                 stripeStatus(value.Status),
		CurrentPeriodStart:     unixTimePtr(periodStart),
		CurrentPeriodEnd:       unixTimePtr(periodEnd),
		TrialEndsAt:            unixTimePtrValue(value.TrialEnd),
		CancelAtPeriodEnd:      value.CancelAtPeriodEnd,
		CanceledAt:             unixTimePtrValue(value.CanceledAt),
		Provider:               stringPtr(providerStripe),
		ProviderCustomerID:     stringPtr(value.Customer),
		ProviderSubscriptionID: stringPtr(value.ID),
		ProviderEventCreatedAt: unixTimePtr(eventCreated),
	})
	return err
}

func stripeStatus(status string) string {
	switch status {
	case "trialing":
		return subscriptions.StatusTrialing
	case "active":
		return subscriptions.StatusActive
	case "canceled", "incomplete_expired":
		return subscriptions.StatusCanceled
	case "incomplete":
		return subscriptions.StatusIncomplete
	case "past_due", "unpaid", "paused":
		return subscriptions.StatusPastDue
	default:
		return subscriptions.StatusPastDue
	}
}

func unixTimePtr(value int64) *time.Time {
	if value <= 0 {
		return nil
	}
	result := time.Unix(value, 0).UTC()
	return &result
}

func unixTimePtrValue(value *int64) *time.Time {
	if value == nil {
		return nil
	}
	return unixTimePtr(*value)
}

func stringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}
