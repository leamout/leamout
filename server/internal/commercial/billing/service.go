package billing

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/commercial/plans"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
	polarintegration "github.com/leamout/leamout/server/internal/integrations/payments/polar"
	"github.com/leamout/leamout/server/pkg/apperror"
)

const providerPolar = "polar"

type Service struct {
	plans         *plans.Service
	subscriptions *subscriptions.Service
	polar         *polarintegration.Client
	productIDs    map[string]string
	frontendURL   string
}

func NewService(
	plansService *plans.Service,
	subscriptionsService *subscriptions.Service,
	polarClient *polarintegration.Client,
	productIDs map[string]string,
	frontendURL ...string,
) *Service {
	var origin string
	if len(frontendURL) > 0 {
		origin = frontendURL[0]
	}
	return &Service{
		plans:         plansService,
		subscriptions: subscriptionsService,
		polar:         polarClient,
		productIDs:    productIDs,
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
	if s.polar == nil || !s.polar.Configured() {
		return SessionResponse{}, apperror.NewServiceUnavailable("Polar billing is not configured", nil)
	}

	plan, err := s.plans.GetByCode(ctx, strings.TrimSpace(req.PlanCode))
	if err != nil {
		return SessionResponse{}, err
	}
	if plan.PricingType != plans.PricingTypeFixed || plan.AmountMinor == nil || *plan.AmountMinor == 0 {
		return SessionResponse{}, apperror.NewBadRequest("plan is not available for Polar checkout")
	}

	if err := s.validateRedirects(req.SuccessURL, req.CancelURL); err != nil {
		return SessionResponse{}, err
	}
	productID := strings.TrimSpace(s.productIDs[plan.Code])
	if productID == "" {
		return SessionResponse{}, apperror.NewServiceUnavailable("Polar product is not configured for this plan", nil)
	}

	existing, err := s.subscriptions.Get(ctx, organizationID)
	if err != nil {
		return SessionResponse{}, err
	}

	if existing != nil {
		if existing.ProviderSubscriptionID != nil &&
			existing.Status != subscriptions.StatusCanceled {
			return SessionResponse{}, apperror.NewConflict("organization already has an active billing subscription")
		}
	}

	session, err := s.polar.CreateCheckoutSession(ctx, polarintegration.CheckoutParams{
		ProductID:      productID,
		OrganizationID: organizationID.String(),
		PlanCode:       plan.Code,
		SuccessURL:     req.SuccessURL,
		ReturnURL:      req.CancelURL,
	})
	if err != nil {
		return SessionResponse{}, apperror.NewServiceUnavailable("create Polar checkout session", err)
	}
	return SessionResponse{
		URL: session.URL,
	}, nil
}

func (s *Service) Portal(
	ctx context.Context,
	organizationID uuid.UUID,
	req PortalRequest,
) (SessionResponse, error) {
	if organizationID == uuid.Nil {
		return SessionResponse{}, apperror.NewBadRequest("organization context required")
	}
	if s.polar == nil || !s.polar.Configured() {
		return SessionResponse{}, apperror.NewServiceUnavailable("Polar billing is not configured", nil)
	}

	existing, err := s.subscriptions.Get(ctx, organizationID)
	if err != nil {
		return SessionResponse{}, err
	}
	if existing == nil ||
		existing.Provider == nil ||
		*existing.Provider != providerPolar ||
		existing.ProviderCustomerID == nil {
		return SessionResponse{}, apperror.NewBadRequest("Polar billing portal is unavailable for this organization")
	}

	if err := s.validateRedirects(req.ReturnURL); err != nil {
		return SessionResponse{}, err
	}
	session, err := s.polar.CreatePortalSession(ctx, polarintegration.PortalParams{
		CustomerID: *existing.ProviderCustomerID,
		ReturnURL:  req.ReturnURL,
	})
	if err != nil {
		return SessionResponse{}, apperror.NewServiceUnavailable("create Polar billing portal session", err)
	}
	return SessionResponse{
		URL: session.URL,
	}, nil
}

func (s *Service) Webhook(ctx context.Context, payload []byte, headers http.Header) error {
	if s.polar == nil || !s.polar.Configured() {
		return apperror.NewServiceUnavailable("Polar billing is not configured", nil)
	}
	event, err := s.polar.ParseWebhook(payload, headers)
	if err != nil {
		return apperror.NewBadRequest("invalid Polar webhook")
	}
	switch event.Type {
	case "subscription.created", "subscription.updated", "subscription.active",
		"subscription.canceled", "subscription.uncanceled", "subscription.cycled",
		"subscription.revoked", "subscription.past_due", "subscription.paused",
		"subscription.resumed", "subscription.migrated":
		var value polarintegration.Subscription
		if err := json.Unmarshal(event.Data, &value); err != nil {
			return apperror.NewBadRequest("invalid Polar subscription event")
		}
		return s.syncPolarSubscription(ctx, event.Timestamp, value)
	default:
		return nil
	}
}

func (s *Service) syncPolarSubscription(
	ctx context.Context,
	eventCreated time.Time,
	value polarintegration.Subscription,
) error {
	organizationID, err := uuid.Parse(strings.TrimSpace(value.Customer.ExternalID))
	if err != nil || organizationID == uuid.Nil {
		return apperror.NewBadRequest("Polar subscription is missing organization metadata")
	}

	if metadataID, ok := value.Metadata["organization_id"]; ok && metadataID != organizationID.String() {
		return apperror.NewBadRequest("Polar subscription organization metadata does not match customer")
	}
	if value.Customer.ID != value.CustomerID {
		return apperror.NewBadRequest("Polar subscription customer identity does not match")
	}
	planCode, err := s.planCodeFromProduct(value)
	if err != nil {
		return err
	}
	plan, err := s.plans.GetByCode(ctx, planCode)
	if err != nil {
		return err
	}

	if value.ID == "" || value.CustomerID == "" || eventCreated.IsZero() {
		return apperror.NewBadRequest("Polar subscription identity or event timestamp is missing")
	}
	periodStart, periodEnd := value.CurrentPeriodStart, value.CurrentPeriodEnd
	if periodStart.IsZero() || !periodEnd.After(periodStart) {
		return apperror.NewBadRequest("Polar subscription period is invalid")
	}
	_, err = s.subscriptions.Sync(ctx, organizationID, subscriptions.SyncRequest{
		PlanID:                 plan.ID,
		Status:                 polarStatus(value.Status),
		CurrentPeriodStart:     &periodStart,
		CurrentPeriodEnd:       &periodEnd,
		TrialEndsAt:            value.TrialEnd,
		CancelAtPeriodEnd:      value.CancelAtPeriodEnd,
		CanceledAt:             value.CanceledAt,
		Provider:               stringPtr(providerPolar),
		ProviderCustomerID:     stringPtr(value.CustomerID),
		ProviderSubscriptionID: stringPtr(value.ID),
		ProviderEventCreatedAt: &eventCreated,
	})
	return err
}

func polarStatus(status string) string {
	switch status {
	case "trialing":
		return subscriptions.StatusTrialing
	case "active":
		return subscriptions.StatusActive
	case "canceled":
		return subscriptions.StatusCanceled
	case "incomplete":
		return subscriptions.StatusIncomplete
	case "past_due", "unpaid", "paused":
		return subscriptions.StatusPastDue
	default:
		return subscriptions.StatusPastDue
	}
}

func stringPtr(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func (s *Service) planCodeFromProduct(value polarintegration.Subscription) (string, error) {
	productID := value.ProductID
	for _, code := range []string{
		"developer",
		"pro",
		"scale",
	} {
		if expected := strings.TrimSpace(s.productIDs[code]); expected != "" && expected == productID {
			return code, nil
		}
	}
	return "", apperror.NewBadRequest("Polar subscription product is not recognized")
}

// validateRedirects prevents authenticated clients from using Polar checkout
// and portal links as open redirects to attacker-controlled websites.
func (s *Service) validateRedirects(values ...string) error {
	trusted, err := url.Parse(strings.TrimSpace(s.frontendURL))
	if err != nil || trusted.Host == "" || (trusted.Scheme != "https" && trusted.Scheme != "http") || trusted.User != nil {
		return apperror.NewServiceUnavailable("billing frontend origin is not configured", nil)
	}
	for _, raw := range values {
		target, parseErr := url.Parse(strings.TrimSpace(raw))
		if parseErr != nil || !target.IsAbs() || target.User != nil ||
			target.Scheme != trusted.Scheme || !strings.EqualFold(target.Host, trusted.Host) ||
			strings.HasPrefix(target.Path, "//") {
			return apperror.NewBadRequest("billing redirect URL must use the configured frontend origin")
		}
	}
	return nil
}

// ReconcileSubscription repairs missed webhooks from authoritative provider state.
// Timestamp the read before the request so a later provider event wins in SQL.
func (s *Service) ReconcileSubscription(ctx context.Context, target ReconciliationTarget) error {
	if s.polar == nil || !s.polar.Configured() {
		return apperror.NewServiceUnavailable("Polar billing is not configured", nil)
	}
	observedAt := time.Now().UTC()
	value, err := s.polar.GetSubscription(ctx, target.SubscriptionID)
	if err != nil {
		return err
	}
	if value.ID != target.SubscriptionID || value.CustomerID != target.CustomerID || value.Customer.ExternalID != target.OrganizationID.String() {
		return apperror.NewConflict("Polar subscription identity does not match local subscription")
	}
	return s.syncPolarSubscription(ctx, observedAt, value)
}
