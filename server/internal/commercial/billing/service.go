package billing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/leamout/leamout/server/internal/commercial/licensing"
	"github.com/leamout/leamout/server/internal/commercial/limits"
	"github.com/leamout/leamout/server/internal/database/pgconv"
	"github.com/leamout/leamout/server/internal/database/sqlc"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/commercial/plans"
	"github.com/leamout/leamout/server/internal/commercial/subscriptions"
	polarintegration "github.com/leamout/leamout/server/internal/integrations/payments/polar"
	"github.com/leamout/leamout/server/pkg/apperror"
)

const providerPolar = "polar"

type Service struct {
	repo          *Repository
	license       *licensing.Service
	limits        *limits.Service
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

	if err := validateRedirects(s.frontendURL, req.SuccessURL, req.CancelURL); err != nil {
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

	if err := validateRedirects(s.frontendURL, req.ReturnURL); err != nil {
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

func (s *Service) SetRepository(repo *Repository) {
	s.repo = repo
}

func (s *Service) SetStatusDependencies(license *licensing.Service, limitsService *limits.Service) {
	s.license = license
	s.limits = limitsService
}

// Verify first, then acknowledge only after durable insertion. Retries retain
// the original receipt; reusing an event ID for different bytes is rejected.
func (s *Service) Webhook(ctx context.Context, payload []byte, headers http.Header) error {
	if s.license.SelfHosted() || s.polar == nil || !s.polar.Configured() || s.repo == nil {
		return apperror.NewServiceUnavailable("Polar webhook inbox is not configured", nil)
	}
	event, err := s.polar.ParseWebhook(payload, headers)
	if err != nil {
		return apperror.NewBadRequest("invalid Polar webhook")
	}
	id := headers.Get("webhook-id")
	if len(id) > 256 || len(event.Type) > 128 {
		return apperror.NewBadRequest("Polar webhook identity is too long")
	}
	digest := sha256.Sum256(payload)
	_, err = s.repo.queries.InsertBillingWebhookEvent(ctx, sqlc.InsertBillingWebhookEventParams{
		Environment:    s.polar.Environment(),
		EventID:        id,
		EventType:      event.Type,
		EventCreatedAt: pgconv.TimeToTimestamptz(event.Timestamp),
		Payload:        payload,
		PayloadHash:    hex.EncodeToString(digest[:]),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return apperror.NewConflict("Polar event ID was reused with different content")
	}
	if err != nil {
		return apperror.NewServiceUnavailable("save Polar webhook receipt", err)
	}
	return nil
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
	var pastDueAt, graceEndsAt *time.Time
	if value.Status == "past_due" {
		existing, err := s.subscriptions.Get(ctx, organizationID)
		if err != nil {
			return err
		}
		if existing != nil &&
			existing.Status == subscriptions.StatusPastDue &&
			existing.ProviderSubscriptionID != nil &&
			*existing.ProviderSubscriptionID == value.ID &&
			existing.PastDueAt != nil &&
			existing.GraceEndsAt != nil {
			pastDueAt, graceEndsAt = existing.PastDueAt, existing.GraceEndsAt
		} else {
			start := eventCreated.UTC()
			if value.PastDueAt != nil {
				start = value.PastDueAt.UTC()
			}
			end := start.AddDate(0, 0, 3)
			pastDueAt, graceEndsAt = &start, &end
		}
	}
	_, err = s.subscriptions.Sync(ctx, organizationID, subscriptions.SyncRequest{
		PlanID:                 plan.ID,
		Status:                 polarStatus(value.Status),
		PastDueAt:              pastDueAt,
		GraceEndsAt:            graceEndsAt,
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
	case "past_due":
		return subscriptions.StatusPastDue
	case "unpaid":
		return subscriptions.StatusCanceled
	case "paused":
		return subscriptions.StatusIncomplete
	default:
		return subscriptions.StatusIncomplete
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

// ReconcileSubscription repairs missed webhooks from authoritative provider state.
// Timestamp the read before the request so a later provider event wins in SQL.
func (s *Service) ReconcileSubscription(ctx context.Context, target ReconciliationTarget) error {
	return s.RecoverOrganization(ctx, target.OrganizationID)
}

// Customer State discovers initial purchases, but excludes past-due customers.
// Always fetch a known subscription as well, including when state is empty.
func (s *Service) RecoverOrganization(ctx context.Context, organizationID uuid.UUID) error {
	if s.license.SelfHosted() || s.polar == nil || !s.polar.Configured() {
		return apperror.NewServiceUnavailable("Polar billing is not configured", nil)
	}
	existing, err := s.subscriptions.Get(ctx, organizationID)
	if err != nil {
		return err
	}
	observedAt := time.Now().UTC()
	state, err := s.polar.GetCustomerStateByExternalID(ctx, organizationID.String())
	if errors.Is(err, polarintegration.ErrNotFound) && existing == nil {
		return nil
	}
	if err != nil {
		return err
	}
	var candidate *polarintegration.CustomerStateSubscription
	for i := range state.ActiveSubscriptions {
		value := &state.ActiveSubscriptions[i]
		candidateSubscription := polarintegration.Subscription{
			ProductID: value.ProductID,
		}
		if _, err := s.planCodeFromProduct(candidateSubscription); err != nil {
			continue
		}
		if candidate != nil && candidate.ID != value.ID {
			return apperror.NewConflict("multiple Leamout subscriptions require review")
		}
		candidate = value
	}
	// First retire/refresh the previous identity. Never replace a live different
	// subscription silently, even if the provider reports another purchase.
	if existing != nil &&
		existing.ProviderSubscriptionID != nil &&
		existing.Provider != nil &&
		*existing.Provider == providerPolar {
		value, err := s.polar.GetSubscription(ctx, *existing.ProviderSubscriptionID)
		if err != nil {
			return err
		}
		if value.Customer.ExternalID != organizationID.String() ||
			value.CustomerID != state.ID ||
			existing.ProviderCustomerID == nil ||
			value.CustomerID != *existing.ProviderCustomerID {
			return apperror.NewConflict("Polar subscription ownership does not match local organization")
		}
		timestamp := observedAt
		if value.ModifiedAt != nil {
			timestamp = value.ModifiedAt.UTC()
		}
		if err := s.syncPolarSubscription(ctx, timestamp, value); err != nil {
			return err
		}
		if candidate == nil || candidate.ID == value.ID {
			return nil
		}
		if value.Status != "canceled" {
			return apperror.NewConflict("active Leamout subscription cannot be replaced")
		}
	}
	if candidate == nil {
		return nil
	}
	value, err := s.polar.GetSubscription(ctx, candidate.ID)
	if err != nil {
		return err
	}
	if value.ID != candidate.ID ||
		value.CustomerID != state.ID ||
		value.Customer.ExternalID != organizationID.String() {
		return apperror.NewConflict("Polar recovered subscription ownership does not match organization")
	}
	timestamp := observedAt
	if value.ModifiedAt != nil {
		timestamp = value.ModifiedAt.UTC()
	}
	return s.syncPolarSubscription(ctx, timestamp, value)
}

func (s *Service) Status(ctx context.Context, organizationID uuid.UUID) (Status, error) {
	if organizationID == uuid.Nil {
		return Status{}, apperror.NewBadRequest("organization context required")
	}
	result := Status{
		DeploymentMode:    licensing.Cloud,
		BillingConfigured: s.polar.Configured(),
	}
	if s.license.SelfHosted() {
		result.DeploymentMode = licensing.SelfHosted
		result.BillingConfigured = false
		status := s.license.Status()
		result.License = &status
		result.Usage.Unlimited = true
		// Self-Hosted Community capacity is operator-managed and unlimited by
		// the commercial licensing system. Enterprise claims gate extensions.
		result.Limits = limits.Effective{}
		return result, nil
	}
	sub, err := s.subscriptions.Get(ctx, organizationID)
	if err != nil {
		return Status{}, err
	}
	planCode := "free"
	plan, err := s.plans.GetByCode(ctx, planCode)
	if err != nil {
		return Status{}, err
	}
	if subscriptions.PaidAccessActive(sub, time.Now().UTC()) {
		plan, err = s.plans.GetByID(ctx, sub.PlanID)
		if err != nil {
			return Status{}, err
		}
	}
	result.Plan = &PlanSummary{
		Code:            plan.Code,
		Name:            plan.Name,
		Currency:        plan.Currency,
		AmountMinor:     plan.AmountMinor,
		BillingInterval: plan.BillingInterval,
	}
	if sub != nil {
		result.Subscription = &SubscriptionSummary{
			Status:            sub.Status,
			CurrentPeriodEnd:  sub.CurrentPeriodEnd,
			CancelAtPeriodEnd: sub.CancelAtPeriodEnd,
			PastDueAt:         sub.PastDueAt,
			GraceEndsAt:       sub.GraceEndsAt,
			Provider:          sub.Provider,
		}
		result.PortalAvailable = result.BillingConfigured &&
			sub.Provider != nil &&
			*sub.Provider == providerPolar &&
			sub.ProviderCustomerID != nil
	}
	result.CheckoutAvailable = result.BillingConfigured &&
		(sub == nil ||
			sub.ProviderSubscriptionID == nil ||
			sub.Status == subscriptions.StatusCanceled)
	if s.limits == nil {
		return Status{}, apperror.NewServiceUnavailable("billing limit resolver is unavailable", nil)
	}
	result.Limits, err = s.limits.Get(ctx, organizationID)
	if err != nil {
		return Status{}, err
	}
	result.Usage, err = s.limits.RuntimeUsage(ctx, organizationID)
	return result, err
}
