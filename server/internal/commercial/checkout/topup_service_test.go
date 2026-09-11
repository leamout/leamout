package checkout

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

type walletStub struct{ wallet wallets.Wallet }

func (s walletStub) Get(context.Context, uuid.UUID, uuid.UUID) (wallets.Wallet, error) {
	return s.wallet, nil
}

type checkoutStub struct {
	created     CreateInput
	checkout    Checkout
	transitions []Transition
}

func (s *checkoutStub) Create(_ context.Context, organizationID uuid.UUID, input CreateInput) (Checkout, error) {
	s.created = input
	s.checkout = Checkout{
		ID: uuid.New(), OrganizationID: organizationID, WalletID: input.WalletID,
		Type: input.Type, Provider: input.Provider, PaymentMethod: input.PaymentMethod,
		Reference: input.Reference, AmountMinor: input.AmountMinor, Currency: input.Currency,
		Status: StatusPending, NextAction: ActionWait, ExpiresAt: input.ExpiresAt,
	}
	return s.checkout, nil
}

func (s *checkoutStub) Get(context.Context, uuid.UUID, uuid.UUID) (Checkout, error) {
	return s.checkout, nil
}

func (s *checkoutStub) ClaimRefresh(_ context.Context, _, _ uuid.UUID, refreshBefore time.Time) (Checkout, error) {
	if s.checkout.UpdatedAt.After(refreshBefore) {
		return Checkout{}, ErrCheckoutNotFound
	}
	s.checkout.UpdatedAt = refreshBefore.Add(10 * time.Second)
	return s.checkout, nil
}

func (s *checkoutStub) Transition(_ context.Context, _ uuid.UUID, _ uuid.UUID, transition Transition) (Checkout, error) {
	s.transitions = append(s.transitions, transition)
	s.checkout.Status = transition.Status
	s.checkout.NextAction = transition.NextAction
	s.checkout.ProviderMessage = transition.ProviderMessage
	return s.checkout, nil
}

type paymentStub struct{ payment commercialpayments.Payment }

func (s *paymentStub) Create(_ context.Context, organizationID uuid.UUID, provider string, input commercialpayments.CreateInput) (commercialpayments.Payment, error) {
	s.payment = commercialpayments.Payment{
		ID: uuid.New(), CheckoutID: input.CheckoutID, OrganizationID: organizationID,
		Provider: provider, Status: input.Status, AmountMinor: input.AmountMinor, Currency: input.Currency,
	}
	return s.payment, nil
}

func (s *paymentStub) GetByCheckout(context.Context, uuid.UUID, uuid.UUID) (commercialpayments.Payment, error) {
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
	event commercialpayments.ProviderEvent
}

func (s *settlementStub) ProcessProviderEvent(_ context.Context, event commercialpayments.ProviderEvent) (TopupSettlement, error) {
	s.event = event
	return TopupSettlement{Applied: true}, nil
}

type providerStub struct {
	request commercialpayments.CheckoutRequest
	event   commercialpayments.ProviderEvent
	payment commercialpayments.ProviderPayment
	session *commercialpayments.CheckoutSession
}

func (s *providerStub) CreateCheckout(_ context.Context, request commercialpayments.CheckoutRequest) (commercialpayments.CheckoutSession, error) {
	s.request = request
	if s.session != nil {
		return *s.session, nil
	}
	return commercialpayments.CheckoutSession{
		Provider: "stripe", ProviderID: "pi_123", Reference: request.Reference,
		ClientSecret: "secret", Status: commercialpayments.StatusPending, NextAction: commercialpayments.NextActionWait,
	}, nil
}

func TestCreateFailsLocalRecordsForMismatchedProviderSession(t *testing.T) {
	organizationID := uuid.New()
	walletID := uuid.New()
	checkouts := &checkoutStub{}
	payments := &paymentStub{}
	provider := &providerStub{session: &commercialpayments.CheckoutSession{
		Provider: "stripe", ProviderID: "pi_123", Reference: "topup.wrong",
	}}
	service := NewTopupService(
		walletStub{wallet: wallets.Wallet{
			ID: walletID, OrganizationID: organizationID, Currency: "USD", Status: wallets.StatusActive,
		}},
		checkouts, payments, &settlementStub{},
		commercialpayments.NewProviderRegistry(map[string]commercialpayments.Provider{"stripe": provider}),
	)

	_, err := service.Create(context.Background(), organizationID, walletID, TopupCreateInput{
		AmountMinor: 2500, Provider: ProviderStripe, Email: "payer@example.com",
	})
	if !errors.Is(err, ErrPaymentMismatch) {
		t.Fatalf("Create() error = %v, want %v", err, ErrPaymentMismatch)
	}
	if payments.payment.Status != commercialpayments.StatusFailed {
		t.Fatalf("payment status = %q, want %q", payments.payment.Status, commercialpayments.StatusFailed)
	}
	if checkouts.checkout.Status != StatusFailed || len(checkouts.transitions) != 1 {
		t.Fatalf("checkout was not failed: %+v", checkouts.checkout)
	}
	transition := checkouts.transitions[0]
	if transition.Expected != StatusPending || transition.NextAction != ActionNone || transition.CompletedAt == nil {
		t.Fatalf("unexpected failure transition: %+v", transition)
	}
}

func (s *providerStub) GetPayment(context.Context, string) (commercialpayments.ProviderPayment, error) {
	return s.payment, nil
}

func (s *providerStub) ParseWebhook([]byte, http.Header) (commercialpayments.ProviderEvent, error) {
	return s.event, nil
}

func TestCreateUsesWalletOwnedPaymentTerms(t *testing.T) {
	organizationID := uuid.New()
	walletID := uuid.New()
	checkouts := &checkoutStub{}
	payments := &paymentStub{}
	provider := &providerStub{}
	service := NewTopupService(
		walletStub{wallet: wallets.Wallet{ID: walletID, OrganizationID: organizationID, Currency: "USD", Status: wallets.StatusActive}},
		checkouts, payments, &settlementStub{},
		commercialpayments.NewProviderRegistry(map[string]commercialpayments.Provider{"stripe": provider}),
	)
	service.now = func() time.Time { return time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC) }

	result, err := service.Create(context.Background(), organizationID, walletID, TopupCreateInput{
		AmountMinor: 2500, Provider: ProviderStripe, Email: " payer@example.com ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if checkouts.created.Currency != "USD" || checkouts.created.AmountMinor != 2500 ||
		checkouts.created.WalletID == nil || *checkouts.created.WalletID != walletID {
		t.Fatalf("checkout terms were not derived from wallet: %+v", checkouts.created)
	}
	if provider.request.Reference != result.Checkout.Reference || provider.request.Currency != "USD" ||
		provider.request.Metadata["wallet_id"] != walletID.String() || provider.request.Email != "payer@example.com" {
		t.Fatalf("provider request mismatch: %+v", provider.request)
	}
	if result.Payment.ProviderID == nil || *result.Payment.ProviderID != "pi_123" || result.Checkout.Status != StatusProcessing {
		t.Fatalf("unexpected checkout result: %+v", result)
	}
}

func TestCreateRejectsPaystackForNonGHSWallet(t *testing.T) {
	provider := &providerStub{}
	service := NewTopupService(
		walletStub{wallet: wallets.Wallet{ID: uuid.New(), Currency: "USD", Status: wallets.StatusActive}},
		&checkoutStub{}, &paymentStub{}, &settlementStub{},
		commercialpayments.NewProviderRegistry(map[string]commercialpayments.Provider{"paystack": provider}),
	)
	_, err := service.Create(context.Background(), uuid.New(), uuid.New(), TopupCreateInput{
		AmountMinor: 100, Provider: ProviderPaystack, Email: "payer@example.com",
		MobileMoney: &commercialpayments.MobileMoney{Phone: "+233200000000", Provider: "mtn"},
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
		walletStub{wallet: wallets.Wallet{ID: walletID, OrganizationID: organizationID, Currency: "GHS", Status: wallets.StatusActive}},
		checkouts, payments, &settlementStub{}, commercialpayments.NewProviderRegistry(map[string]commercialpayments.Provider{"paystack": provider}),
	)

	_, err := service.Create(context.Background(), organizationID, walletID, TopupCreateInput{
		AmountMinor: 2500, Provider: ProviderPaystack, Email: "payer@example.com",
		MobileMoney: &commercialpayments.MobileMoney{Phone: "0240000000", Provider: "mtn"},
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

func (s *referenceProviderStub) CreateCheckout(_ context.Context, request commercialpayments.CheckoutRequest) (commercialpayments.CheckoutSession, error) {
	s.request = request
	return commercialpayments.CheckoutSession{Provider: s.provider, Reference: request.Reference, Status: commercialpayments.StatusProcessing}, nil
}

func TestGetReconcilesMaturePaystackCharge(t *testing.T) {
	now := time.Date(2026, 9, 10, 1, 0, 0, 0, time.UTC)
	organizationID := uuid.New()
	walletID := uuid.New()
	checkoutID := uuid.New()
	checkouts := &checkoutStub{checkout: Checkout{
		ID: checkoutID, OrganizationID: organizationID, WalletID: &walletID,
		Provider: ProviderPaystack, Reference: "topup.123", AmountMinor: 2500,
		Currency: "GHS", Status: StatusProcessing, UpdatedAt: now.Add(-11 * time.Second),
	}}
	payments := &paymentStub{payment: commercialpayments.Payment{ID: uuid.New(), CheckoutID: checkoutID}}
	provider := &providerStub{payment: commercialpayments.ProviderPayment{
		Provider: "paystack", ProviderID: "42", Reference: "topup.123",
		AmountMinor: 2500, Currency: "GHS", Status: commercialpayments.StatusSucceeded,
	}}
	settlements := &settlementStub{}
	service := NewTopupService(nil, checkouts, payments, settlements, commercialpayments.NewProviderRegistry(map[string]commercialpayments.Provider{"paystack": provider}))
	service.now = func() time.Time { return now }

	if _, err := service.Get(context.Background(), organizationID, checkoutID); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if settlements.event.Type != "charge.success" || settlements.event.Payment.ProviderID != "42" {
		t.Fatalf("charge lookup was not reconciled: %+v", settlements.event)
	}
}
