package topups

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/checkout"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/commercial/wallets"
	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
)

type walletStub struct{ wallet wallets.Wallet }

func (s walletStub) Get(context.Context, uuid.UUID, uuid.UUID) (wallets.Wallet, error) {
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
		ID: uuid.New(), CheckoutOrderID: input.CheckoutOrderID, OrganizationID: organizationID,
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

func (s *settlementStub) Reconcile(_ context.Context, event paymentprovider.Event) (Settlement, error) {
	s.event = event
	return Settlement{Applied: true}, nil
}

type providerStub struct {
	request paymentprovider.CheckoutRequest
	event   paymentprovider.Event
}

func (s *providerStub) CreateCheckout(_ context.Context, request paymentprovider.CheckoutRequest) (paymentprovider.CheckoutSession, error) {
	s.request = request
	return paymentprovider.CheckoutSession{
		Provider: "stripe", ProviderID: "pi_123", Reference: request.Reference,
		ClientSecret: "secret", Status: paymentprovider.StatusPending, NextAction: paymentprovider.NextActionWait,
	}, nil
}

func (s *providerStub) GetPayment(context.Context, string) (paymentprovider.Payment, error) {
	return paymentprovider.Payment{}, nil
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
	service := NewService(
		walletStub{wallet: wallets.Wallet{ID: walletID, OrganizationID: organizationID, Currency: "USD", Status: wallets.StatusActive}},
		checkouts, payments, &settlementStub{},
		map[string]paymentprovider.Provider{"stripe": provider},
	)
	service.now = func() time.Time { return time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC) }

	result, err := service.Create(context.Background(), organizationID, walletID, CreateInput{
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
	service := NewService(
		walletStub{wallet: wallets.Wallet{ID: uuid.New(), Currency: "USD", Status: wallets.StatusActive}},
		&checkoutStub{}, &paymentStub{}, &settlementStub{},
		map[string]paymentprovider.Provider{"paystack": provider},
	)
	_, err := service.Create(context.Background(), uuid.New(), uuid.New(), CreateInput{
		AmountMinor: 100, Provider: checkout.ProviderPaystack, Email: "payer@example.com",
		MobileMoney: &paymentprovider.MobileMoney{Phone: "+233200000000", Provider: "mtn"},
	})
	if err != ErrInvalidTopup {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalidTopup)
	}
	if provider.request.Reference != "" {
		t.Fatal("provider must not be called for an invalid wallet currency")
	}
}

func TestWebhookPassesOnlyAuthenticatedProviderEventToSettlement(t *testing.T) {
	event := paymentprovider.Event{
		Provider: "stripe", ProviderEventID: "evt_123", Type: "payment_intent.succeeded",
		Payment: paymentprovider.Payment{Reference: "topup.123", Status: paymentprovider.StatusSucceeded},
		Raw:     []byte(`{"id":"evt_123"}`),
	}
	provider := &providerStub{event: event}
	settlements := &settlementStub{}
	service := NewService(nil, nil, nil, settlements, map[string]paymentprovider.Provider{"stripe": provider})

	result, err := service.Webhook(context.Background(), "stripe", []byte(`{}`), http.Header{})
	if err != nil {
		t.Fatalf("Webhook() error = %v", err)
	}
	if !result.Applied || settlements.event.ProviderEventID != "evt_123" {
		t.Fatalf("verified event was not reconciled: %+v", settlements.event)
	}
}
