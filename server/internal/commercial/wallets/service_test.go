package wallets

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/checkout"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
)

type walletStub struct{ wallet Wallet }

func (s walletStub) Get(context.Context, uuid.UUID, uuid.UUID) (Wallet, error) {
	return s.wallet, nil
}

type checkoutStub struct {
	created     checkout.CreateInput
	order       checkout.Order
	transitions []checkout.Transition
}

func (s *checkoutStub) Create(_ context.Context, organizationID uuid.UUID, input checkout.CreateInput) (checkout.Order, error) {
	s.created = input
	s.order = checkout.Order{
		ID: uuid.New(), OrganizationID: organizationID, WalletID: input.WalletID,
		Type: input.Type, Provider: input.Provider, PaymentMethod: input.PaymentMethod,
		Reference: input.Reference, AmountMinor: input.AmountMinor, Currency: input.Currency,
		Status: checkout.StatusPending, NextAction: checkout.ActionWait, ExpiresAt: input.ExpiresAt,
	}
	return s.order, nil
}

func (s *checkoutStub) Get(context.Context, uuid.UUID, uuid.UUID) (checkout.Order, error) {
	return s.order, nil
}

func (s *checkoutStub) ClaimRefresh(_ context.Context, _, _ uuid.UUID, refreshBefore time.Time) (checkout.Order, error) {
	if s.order.UpdatedAt.After(refreshBefore) {
		return checkout.Order{}, checkout.ErrOrderNotFound
	}
	s.order.UpdatedAt = refreshBefore.Add(10 * time.Second)
	return s.order, nil
}

func (s *checkoutStub) Transition(_ context.Context, _ uuid.UUID, _ uuid.UUID, transition checkout.Transition) (checkout.Order, error) {
	s.transitions = append(s.transitions, transition)
	s.order.Status = transition.Status
	s.order.NextAction = transition.NextAction
	s.order.ProviderMessage = transition.ProviderMessage
	return s.order, nil
}

type paymentStub struct{ payment commercialpayments.Payment }

func (s *paymentStub) Create(_ context.Context, organizationID uuid.UUID, provider string, input commercialpayments.CreateInput) (commercialpayments.Payment, error) {
	s.payment = commercialpayments.Payment{
		ID: uuid.New(), CheckoutID: input.CheckoutID, OrganizationID: organizationID,
		Provider: provider, Status: input.Status, AmountMinor: input.AmountMinor, Currency: input.Currency,
	}
	return s.payment, nil
}

func (s *paymentStub) GetByCheckoutOrder(context.Context, uuid.UUID, uuid.UUID) (commercialpayments.Payment, error) {
	return s.payment, nil
}

func (s *paymentStub) SetProviderID(_ context.Context, _ uuid.UUID, _ uuid.UUID, providerID string, status commercialpayments.Status) (commercialpayments.Payment, error) {
	s.payment.ProviderID = &providerID
	s.payment.Status = status
	return s.payment, nil
}

func (s *paymentStub) UpdateStatus(_ context.Context, _ uuid.UUID, _ uuid.UUID, status commercialpayments.Status, _ *time.Time) (commercialpayments.Payment, error) {
	s.payment.Status = status
	return s.payment, nil
}

type settlementStub struct {
	event paymentprovider.Event
}

func (s *settlementStub) Reconcile(_ context.Context, event paymentprovider.Event) (TopupSettlement, error) {
	s.event = event
	return TopupSettlement{Applied: true}, nil
}

type providerStub struct {
	request paymentprovider.CheckoutRequest
	event   paymentprovider.Event
	payment paymentprovider.Payment
	session *paymentprovider.CheckoutSession
}

func (s *providerStub) CreateCheckout(_ context.Context, request paymentprovider.CheckoutRequest) (paymentprovider.CheckoutSession, error) {
	s.request = request
	if s.session != nil {
		return *s.session, nil
	}
	return paymentprovider.CheckoutSession{
		Provider: "stripe", ProviderID: "pi_123", Reference: request.Reference,
		ClientSecret: "secret", Status: paymentprovider.StatusPending, NextAction: paymentprovider.NextActionWait,
	}, nil
}

func TestCreateFailsLocalRecordsForMismatchedProviderSession(t *testing.T) {
	organizationID := uuid.New()
	walletID := uuid.New()
	checkouts := &checkoutStub{}
	payments := &paymentStub{}
	provider := &providerStub{session: &paymentprovider.CheckoutSession{
		Provider: "stripe", ProviderID: "pi_123", Reference: "topup.wrong",
	}}
	service := NewTopupService(
		walletStub{wallet: Wallet{
			ID: walletID, OrganizationID: organizationID, Currency: "USD", Status: StatusActive,
		}},
		checkouts, payments, &settlementStub{},
		map[string]paymentprovider.Provider{"stripe": provider},
	)

	_, err := service.Create(context.Background(), organizationID, walletID, TopupCreateInput{
		AmountMinor: 2500, Provider: checkout.ProviderStripe, Email: "payer@example.com",
	})
	if !errors.Is(err, ErrPaymentMismatch) {
		t.Fatalf("Create() error = %v, want %v", err, ErrPaymentMismatch)
	}
	if payments.payment.Status != commercialpayments.StatusFailed {
		t.Fatalf("payment status = %q, want %q", payments.payment.Status, commercialpayments.StatusFailed)
	}
	if checkouts.order.Status != checkout.StatusFailed || len(checkouts.transitions) != 1 {
		t.Fatalf("checkout was not failed: %+v", checkouts.order)
	}
	transition := checkouts.transitions[0]
	if transition.Expected != checkout.StatusPending || transition.NextAction != checkout.ActionNone || transition.CompletedAt == nil {
		t.Fatalf("unexpected failure transition: %+v", transition)
	}
}

func (s *providerStub) GetPayment(context.Context, string) (paymentprovider.Payment, error) {
	return s.payment, nil
}

func (s *providerStub) ParseWebhook([]byte, http.Header) (paymentprovider.Event, error) {
	return s.event, nil
}

func TestCreateUsesWalletOwnedPaymentTerms(t *testing.T) {
	organizationID := uuid.New()
	walletID := uuid.New()
	checkouts := &checkoutStub{}
	payments := &paymentStub{}
	provider := &providerStub{}
	service := NewTopupService(
		walletStub{wallet: Wallet{ID: walletID, OrganizationID: organizationID, Currency: "USD", Status: StatusActive}},
		checkouts, payments, &settlementStub{},
		map[string]paymentprovider.Provider{"stripe": provider},
	)
	service.now = func() time.Time { return time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC) }

	result, err := service.Create(context.Background(), organizationID, walletID, TopupCreateInput{
		AmountMinor: 2500, Provider: checkout.ProviderStripe, Email: " payer@example.com ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if checkouts.created.Currency != "USD" || checkouts.created.AmountMinor != 2500 ||
		checkouts.created.WalletID == nil || *checkouts.created.WalletID != walletID {
		t.Fatalf("checkout terms were not derived from wallet: %+v", checkouts.created)
	}
	if provider.request.Reference != result.Order.Reference || provider.request.Currency != "USD" ||
		provider.request.Metadata["wallet_id"] != walletID.String() || provider.request.Email != "payer@example.com" {
		t.Fatalf("provider request mismatch: %+v", provider.request)
	}
	if result.Payment.ProviderID == nil || *result.Payment.ProviderID != "pi_123" || result.Order.Status != checkout.StatusProcessing {
		t.Fatalf("unexpected checkout result: %+v", result)
	}
}

func TestCreateRejectsPaystackForNonGHSWallet(t *testing.T) {
	provider := &providerStub{}
	service := NewTopupService(
		walletStub{wallet: Wallet{ID: uuid.New(), Currency: "USD", Status: StatusActive}},
		&checkoutStub{}, &paymentStub{}, &settlementStub{},
		map[string]paymentprovider.Provider{"paystack": provider},
	)
	_, err := service.Create(context.Background(), uuid.New(), uuid.New(), TopupCreateInput{
		AmountMinor: 100, Provider: checkout.ProviderPaystack, Email: "payer@example.com",
		MobileMoney: &paymentprovider.MobileMoney{Phone: "+233200000000", Provider: "mtn"},
	})
	if !errors.Is(err, ErrInvalidTopup) {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalidTopup)
	}
	if provider.request.Reference != "" {
		t.Fatal("provider must not be called for an invalid wallet currency")
	}
}

func TestCreateAcceptsPaystackChargeWithoutTransactionID(t *testing.T) {
	organizationID := uuid.New()
	walletID := uuid.New()
	payments := &paymentStub{}
	checkouts := &checkoutStub{}
	provider := &referenceProviderStub{provider: "paystack"}
	service := NewTopupService(
		walletStub{wallet: Wallet{ID: walletID, OrganizationID: organizationID, Currency: "GHS", Status: StatusActive}},
		checkouts, payments, &settlementStub{}, map[string]paymentprovider.Provider{"paystack": provider},
	)

	_, err := service.Create(context.Background(), organizationID, walletID, TopupCreateInput{
		AmountMinor: 2500, Provider: checkout.ProviderPaystack, Email: "payer@example.com",
		MobileMoney: &paymentprovider.MobileMoney{Phone: "0240000000", Provider: "mtn"},
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if payments.payment.ProviderID != nil || payments.payment.Status != commercialpayments.StatusProcessing {
		t.Fatalf("payment = %+v", payments.payment)
	}
}

type referenceProviderStub struct {
	providerStub
	provider string
}

func (s *referenceProviderStub) CreateCheckout(_ context.Context, request paymentprovider.CheckoutRequest) (paymentprovider.CheckoutSession, error) {
	s.request = request
	return paymentprovider.CheckoutSession{Provider: s.provider, Reference: request.Reference, Status: paymentprovider.StatusProcessing}, nil
}

func TestWebhookPassesOnlyAuthenticatedProviderEventToSettlement(t *testing.T) {
	event := paymentprovider.Event{
		Provider: "stripe", ProviderEventID: "evt_123", Type: "checkout.session.completed",
		Payment: paymentprovider.Payment{Reference: "topup.123", Status: paymentprovider.StatusSucceeded},
		Raw:     []byte(`{"id":"evt_123"}`),
	}
	provider := &providerStub{event: event}
	settlements := &settlementStub{}
	service := NewTopupService(nil, nil, nil, settlements, map[string]paymentprovider.Provider{"stripe": provider})

	result, err := service.Webhook(context.Background(), "stripe", []byte(`{}`), http.Header{})
	if err != nil {
		t.Fatalf("Webhook() error = %v", err)
	}
	if !result.Applied || settlements.event.ProviderEventID != "evt_123" {
		t.Fatalf("verified event was not reconciled: %+v", settlements.event)
	}
}

func TestWebhookIgnoresAuthenticatedNonPaymentEvent(t *testing.T) {
	provider := &providerStub{event: paymentprovider.Event{
		Provider: "paystack", ProviderEventID: "refund:1", Type: "refund.processed",
		Payment: paymentprovider.Payment{Status: paymentprovider.StatusSucceeded}, Raw: []byte(`{"event":"refund.processed"}`),
	}}
	settlements := &settlementStub{}
	service := NewTopupService(nil, nil, nil, settlements, map[string]paymentprovider.Provider{"paystack": provider})

	result, err := service.Webhook(context.Background(), "paystack", []byte(`{}`), http.Header{})
	if err != nil || result.Applied || settlements.event.ProviderEventID != "" {
		t.Fatalf("non-payment event was reconciled: result=%+v event=%+v err=%v", result, settlements.event, err)
	}
}

func TestGetReconcilesMaturePaystackCharge(t *testing.T) {
	now := time.Date(2026, 9, 10, 1, 0, 0, 0, time.UTC)
	organizationID := uuid.New()
	walletID := uuid.New()
	orderID := uuid.New()
	checkouts := &checkoutStub{order: checkout.Order{
		ID: orderID, OrganizationID: organizationID, WalletID: &walletID,
		Provider: checkout.ProviderPaystack, Reference: "topup.123", AmountMinor: 2500,
		Currency: "GHS", Status: checkout.StatusProcessing, UpdatedAt: now.Add(-11 * time.Second),
	}}
	payments := &paymentStub{payment: commercialpayments.Payment{ID: uuid.New(), CheckoutID: orderID}}
	provider := &providerStub{payment: paymentprovider.Payment{
		Provider: "paystack", ProviderID: "42", Reference: "topup.123",
		AmountMinor: 2500, Currency: "GHS", Status: paymentprovider.StatusSucceeded,
	}}
	settlements := &settlementStub{}
	service := NewTopupService(nil, checkouts, payments, settlements, map[string]paymentprovider.Provider{"paystack": provider})
	service.now = func() time.Time { return now }

	if _, err := service.Get(context.Background(), organizationID, orderID); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if settlements.event.Type != "charge.success" || settlements.event.Payment.ProviderID != "42" {
		t.Fatalf("charge lookup was not reconciled: %+v", settlements.event)
	}
}
