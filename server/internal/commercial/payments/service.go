package payments

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type paymentStore interface {
	Create(context.Context, uuid.UUID, string, CreateInput) (Payment, error)
	GetByCheckout(context.Context, uuid.UUID, uuid.UUID) (Payment, error)
	SetProviderID(context.Context, uuid.UUID, uuid.UUID, string, Status) (Payment, error)
	UpdateStatus(context.Context, uuid.UUID, uuid.UUID, Status, *time.Time) (Payment, error)
}

type eventStore interface {
	processProviderEvent(context.Context, ProviderEvent) (Settlement, error)
	checkProviderEventReplay(context.Context, ProviderEvent) error
}

type Service struct {
	payments  paymentStore
	events    eventStore
	providers *ProviderRegistry
}

type StartInput struct {
	CheckoutID  uuid.UUID
	Provider    string
	Reference   string
	AmountMinor int64
	Currency    string
	Email       string
	CallbackURL string
	Metadata    map[string]string
	MobileMoney *MobileMoney
}

type StartResult struct {
	Payment Payment
	Session CheckoutSession
}

type ContinueInput struct {
	Provider  string
	Reference string
	Action    NextAction
	Value     string
}

func NewService(repository *Repository, providers *ProviderRegistry) *Service {
	return newService(repository, repository, providers)
}

func newService(payments paymentStore, events eventStore, providers *ProviderRegistry) *Service {
	if providers == nil {
		providers = NewProviderRegistry()
	}
	return &Service{payments: payments, events: events, providers: providers}
}

func (s *Service) ProviderAvailable(name string) bool {
	if s == nil || s.providers == nil {
		return false
	}
	_, ok := s.providers.Get(name)
	return ok
}

func (s *Service) GetByCheckout(ctx context.Context, organizationID, checkoutID uuid.UUID) (Payment, error) {
	if s == nil || s.payments == nil {
		return Payment{}, ErrPaymentNotFound
	}
	return s.payments.GetByCheckout(ctx, organizationID, checkoutID)
}

func (s *Service) Start(ctx context.Context, organizationID uuid.UUID, input StartInput) (StartResult, error) {
	if s == nil || s.payments == nil || organizationID == uuid.Nil || input.CheckoutID == uuid.Nil || input.Provider == "" {
		return StartResult{}, ErrInvalidPayment
	}

	provider, ok := s.providers.Get(input.Provider)
	if !ok {
		return StartResult{}, ErrInvalidPayment
	}

	request, err := ValidateCheckout(CheckoutRequest{
		Reference:   input.Reference,
		AmountMinor: input.AmountMinor,
		Currency:    input.Currency,
		Email:       input.Email,
		CallbackURL: input.CallbackURL,
		Metadata:    input.Metadata,
		MobileMoney: input.MobileMoney,
	})
	if err != nil {
		return StartResult{}, ErrInvalidPayment
	}

	payment, err := s.payments.Create(ctx, organizationID, input.Provider, CreateInput{
		CheckoutID:  input.CheckoutID,
		Status:      StatusPending,
		AmountMinor: request.AmountMinor,
		Currency:    request.Currency,
	})
	if err != nil {
		return StartResult{}, err
	}

	session, err := provider.CreateCheckout(ctx, request)
	if err != nil {
		s.failPayment(ctx, organizationID, payment.ID)
		return StartResult{}, err
	}
	if !validProviderSession(input.Provider, request.Reference, session) {
		s.failPayment(ctx, organizationID, payment.ID)
		return StartResult{}, ErrPaymentMismatch
	}

	if session.ProviderID == "" {
		payment, err = s.payments.UpdateStatus(ctx, organizationID, payment.ID, StatusProcessing, nil)
	} else {
		payment, err = s.payments.SetProviderID(ctx, organizationID, payment.ID, session.ProviderID, StatusProcessing)
	}
	if err != nil {
		return StartResult{}, err
	}

	return StartResult{Payment: payment, Session: session}, nil
}

func (s *Service) Continue(ctx context.Context, input ContinueInput) (CheckoutSession, error) {
	if s == nil || s.providers == nil || input.Provider == "" || input.Reference == "" {
		return CheckoutSession{}, ErrInvalidPayment
	}
	provider, ok := s.providers.Get(input.Provider)
	if !ok {
		return CheckoutSession{}, ErrInvalidPayment
	}
	continuation, ok := provider.(ContinuationProvider)
	if !ok {
		return CheckoutSession{}, ErrInvalidPayment
	}
	return continuation.ContinueCheckout(ctx, ContinueCheckoutRequest{
		Reference: input.Reference,
		Action:    input.Action,
		Value:     input.Value,
	})
}

func (s *Service) Refresh(ctx context.Context, providerName, reference string) (Settlement, bool, error) {
	if s == nil || s.providers == nil || providerName == "" || reference == "" {
		return Settlement{}, false, nil
	}
	provider, ok := s.providers.Get(providerName)
	if !ok {
		return Settlement{}, false, nil
	}

	payment, err := provider.GetPayment(ctx, reference)
	if err != nil {
		return Settlement{}, false, nil
	}
	eventType, ok := lookupEventType(providerName, payment.Status)
	if !ok {
		return Settlement{}, false, nil
	}

	raw, _ := json.Marshal(map[string]string{
		"source":    "payment_lookup",
		"reference": reference,
		"status":    string(payment.Status),
	})
	event := ProviderEvent{
		Provider:        providerName,
		ProviderEventID: "lookup:" + reference + ":" + string(payment.Status),
		Type:            eventType,
		Payment:         payment,
		Raw:             raw,
	}
	settlement, err := s.ProcessProviderEvent(ctx, event)
	if err != nil {
		return Settlement{}, true, err
	}
	return settlement, true, nil
}

func (s *Service) ParseWebhook(providerName string, payload []byte, headers http.Header) (ProviderEvent, error) {
	if s == nil || s.providers == nil {
		return ProviderEvent{}, ErrPaymentMismatch
	}
	provider, ok := s.providers.Get(providerName)
	if !ok {
		return ProviderEvent{}, ErrPaymentMismatch
	}
	event, err := provider.ParseWebhook(payload, headers)
	if err != nil || event.Provider != providerName {
		return ProviderEvent{}, ErrPaymentMismatch
	}
	return event, nil
}

// ProcessProviderEvent owns normalized payment-provider event validation and
// reconciliation. Provider adapters authenticate and parse payloads before
// they reach this boundary.
func (s *Service) ProcessProviderEvent(ctx context.Context, event ProviderEvent) (Settlement, error) {
	if s == nil || s.events == nil {
		return Settlement{}, ErrPaymentMismatch
	}
	relevant, valid := classifyProviderEvent(event)
	if !valid {
		return Settlement{}, ErrPaymentMismatch
	}
	if !relevant {
		return Settlement{}, s.events.checkProviderEventReplay(ctx, event)
	}
	return s.events.processProviderEvent(ctx, event)
}

func (s *Service) failPayment(ctx context.Context, organizationID, paymentID uuid.UUID) {
	if s == nil || s.payments == nil || paymentID == uuid.Nil {
		return
	}
	_, _ = s.payments.UpdateStatus(ctx, organizationID, paymentID, StatusFailed, nil)
}
