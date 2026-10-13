package billing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
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
	frontendURL   string
}

func NewService(
	plansService *plans.Service,
	subscriptionsService *subscriptions.Service,
	stripeClient *stripeintegration.Client,
	priceIDs map[string]string,
	frontendURL ...string,
) *Service {
	var origin string
	if len(frontendURL)>0 {origin=frontendURL[0]}
	return &Service{
		plans:         plansService,
		subscriptions: subscriptionsService,
		stripe:        stripeClient,
		priceIDs:      priceIDs,
		frontendURL:   origin,
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

	if err := s.validateRedirects(req.SuccessURL, req.CancelURL);err!=nil{
		return SessionResponse{},err
	}
	if strings.TrimSpace(req.IdempotencyKey)==""{
		return SessionResponse{},apperror.NewBadRequest("Idempotency-Key is required for checkout")
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

	digest := sha256.Sum256([]byte(organizationID.String()+":"+req.IdempotencyKey))
	session, err := s.stripe.CreateCheckoutSession(ctx, stripeintegration.CheckoutParams{
		IdempotencyKey: "leamout-checkout-"+hex.EncodeToString(digest[:]),
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

	if err := s.validateRedirects(req.ReturnURL);err!=nil {
		return SessionResponse{},err
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

	planCode, err := s.planCodeFromPrice(value)
	if err != nil {return err}
	plan, err := s.plans.GetByCode(ctx, planCode)
	if err != nil {return err}

	if value.ID=="" || value.Customer=="" || eventCreated<=0 {
		return apperror.NewBadRequest("Stripe subscription identity or event timestamp is missing")
	}
	periodStart, periodEnd := value.CurrentPeriod()
	if periodStart<=0 || periodEnd<=periodStart {
		return apperror.NewBadRequest("Stripe subscription period is invalid")
	}
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

func (s *Service) planCodeFromPrice(value stripeintegration.Subscription) (string,error) {
	if len(value.Items.Data)!=1 {
		return "",apperror.NewBadRequest("Stripe subscription must contain exactly one price")
	}
	priceID:=strings.TrimSpace(value.Items.Data[0].Price.ID)
	if priceID=="" {
		return "",apperror.NewBadRequest("Stripe subscription price is missing")
	}
	for _,code:=range []string{"developer","pro","scale"} {
		if expected:=strings.TrimSpace(s.priceIDs[code]);expected!="" && expected==priceID {
			return code,nil
		}
	}
	return "",apperror.NewBadRequest("Stripe subscription price is not recognized")
}

// validateRedirects prevents authenticated clients from using Stripe checkout
// and portal links as open redirects to attacker-controlled websites.
func (s *Service) validateRedirects(values ...string) error {
	trusted,err:=url.Parse(strings.TrimSpace(s.frontendURL))
	if err!=nil || trusted.Host=="" || (trusted.Scheme!="https" && trusted.Scheme!="http") || trusted.User!=nil {
		return apperror.NewServiceUnavailable("billing frontend origin is not configured",nil)
	}
	for _,raw:=range values {
		target,parseErr:=url.Parse(strings.TrimSpace(raw))
		if parseErr!=nil || !target.IsAbs() || target.User!=nil ||
			target.Scheme!=trusted.Scheme || !strings.EqualFold(target.Host,trusted.Host) ||
			strings.HasPrefix(target.Path,"//") {
			return apperror.NewBadRequest("billing redirect URL must use the configured frontend origin")
		}
	}
	return nil
}
