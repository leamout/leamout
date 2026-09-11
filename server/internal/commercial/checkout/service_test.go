package checkout

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/catalog"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

type checkoutStoreStub struct {
	checkout     Checkout
	createdInput CreateInput
	startedInput StartPayment
	transition   Transition
	creates      int
	starts       int
	transitions  int
}

func (s *checkoutStoreStub) Create(_ context.Context, organizationID uuid.UUID, input CreateInput) (Checkout, error) {
	s.creates++
	s.createdInput = input
	result := s.checkout
	if result.ID == uuid.Nil {
		result.ID = uuid.New()
	}
	result.OrganizationID = organizationID
	result.WalletID = input.WalletID
	result.PriceID = input.PriceID
	result.Type = input.Type
	result.Reference = input.Reference
	result.AmountMinor = input.AmountMinor
	result.Currency = input.Currency
	result.Status = StatusPending
	result.NextAction = ActionWait
	result.ExpiresAt = input.ExpiresAt
	result.Metadata = input.Metadata
	s.checkout = result
	return result, nil
}

func (s *checkoutStoreStub) StartPayment(_ context.Context, _, _ uuid.UUID, input StartPayment) (Checkout, error) {
	s.starts++
	s.startedInput = input
	s.checkout.Provider = input.Provider
	s.checkout.PaymentMethod = input.PaymentMethod
	s.checkout.Status = StatusProcessing
	s.checkout.NextAction = ActionWait
	return s.checkout, nil
}

func (s *checkoutStoreStub) Get(context.Context, uuid.UUID, uuid.UUID) (Checkout, error) {
	return s.checkout, nil
}

func (s *checkoutStoreStub) GetByReference(context.Context, string) (Checkout, error) {
	return s.checkout, nil
}

func (s *checkoutStoreStub) Transition(_ context.Context, _, _ uuid.UUID, input Transition) (Checkout, error) {
	s.transitions++
	s.transition = input
	s.checkout.Status = input.Status
	s.checkout.NextAction = input.NextAction
	s.checkout.ProviderMessage = input.ProviderMessage
	s.checkout.CompletedAt = input.CompletedAt
	return s.checkout, nil
}

func (s *checkoutStoreStub) ClaimRefresh(context.Context, uuid.UUID, uuid.UUID, time.Time) (Checkout, error) {
	return Checkout{}, ErrCheckoutNotFound
}

func (s *checkoutStoreStub) Expire(context.Context) ([]Checkout, error) { return nil, nil }

type walletServiceStub struct {
	wallet    wallets.Wallet
	postInput wallets.PostEntryInput
	postErr   error
	posts     int
}

func (s *walletServiceStub) Get(context.Context, uuid.UUID, uuid.UUID) (wallets.Wallet, error) {
	return s.wallet, nil
}

func (s *walletServiceStub) Post(
	_ context.Context,
	_, _ uuid.UUID,
	input wallets.PostEntryInput,
) (wallets.LedgerEntry, error) {
	s.posts++
	s.postInput = input
	return wallets.LedgerEntry{}, s.postErr
}

type catalogServiceStub struct {
	price   catalog.Price
	plan    catalog.Plan
	product catalog.Product
}

func (s *catalogServiceStub) GetPrice(context.Context, uuid.UUID) (catalog.Price, error) {
	return s.price, nil
}

func (s *catalogServiceStub) GetPlan(context.Context, uuid.UUID) (catalog.Plan, error) {
	return s.plan, nil
}

func (s *catalogServiceStub) GetProduct(context.Context, uuid.UUID) (catalog.Product, error) {
	return s.product, nil
}

type subscriptionServiceStub struct {
	current    subscriptions.Subscription
	currentErr error
	created    subscriptions.CreateInput
	createErr  error
	creates    int
}

func (s *subscriptionServiceStub) Create(
	_ context.Context,
	_ uuid.UUID,
	input subscriptions.CreateInput,
) (subscriptions.Subscription, error) {
	s.creates++
	s.created = input
	return subscriptions.Subscription{}, s.createErr
}

func (s *subscriptionServiceStub) Current(context.Context, uuid.UUID) (subscriptions.Subscription, error) {
	if s.currentErr != nil {
		return subscriptions.Subscription{}, s.currentErr
	}
	return s.current, nil
}

type paymentStoreStub struct {
	payment  commercialpayments.Payment
	created  commercialpayments.CreateInput
	provider string
}

func (s *paymentStoreStub) Create(
	_ context.Context,
	organizationID uuid.UUID,
	provider string,
	input commercialpayments.CreateInput,
) (commercialpayments.Payment, error) {
	s.created = input
	s.provider = provider
	result := s.payment
	if result.ID == uuid.Nil {
		result.ID = uuid.New()
	}
	result.OrganizationID = organizationID
	result.CheckoutID = input.CheckoutID
	result.Provider = provider
	result.Status = input.Status
	result.AmountMinor = input.AmountMinor
	result.Currency = input.Currency
	s.payment = result
	return result, nil
}

func (s *paymentStoreStub) GetByCheckout(context.Context, uuid.UUID, uuid.UUID) (commercialpayments.Payment, error) {
	if s.payment.ID == uuid.Nil {
		return commercialpayments.Payment{}, commercialpayments.ErrPaymentNotFound
	}
	return s.payment, nil
}

func (s *paymentStoreStub) SetProviderID(
	_ context.Context,
	_, _ uuid.UUID,
	providerID string,
	status commercialpayments.Status,
) (commercialpayments.Payment, error) {
	s.payment.ProviderID = &providerID
	s.payment.Status = status
	return s.payment, nil
}

func (s *paymentStoreStub) UpdateStatus(
	_ context.Context,
	_, _ uuid.UUID,
	status commercialpayments.Status,
	_ *time.Time,
) (commercialpayments.Payment, error) {
	s.payment.Status = status
	return s.payment, nil
}

type paymentEventsStub struct{}

func (paymentEventsStub) ProcessProviderEvent(
	context.Context,
	commercialpayments.ProviderEvent,
) (commercialpayments.Settlement, error) {
	return commercialpayments.Settlement{}, nil
}

type providerStub struct {
	session commercialpayments.CheckoutSession
}

func (s providerStub) CreateCheckout(context.Context, commercialpayments.CheckoutRequest) (commercialpayments.CheckoutSession, error) {
	return s.session, nil
}

func (providerStub) GetPayment(context.Context, string) (commercialpayments.ProviderPayment, error) {
	return commercialpayments.ProviderPayment{}, nil
}

func (providerStub) ParseWebhook([]byte, http.Header) (commercialpayments.ProviderEvent, error) {
	return commercialpayments.ProviderEvent{}, nil
}

func TestServiceRejectsInvalidCheckoutBeforePersistence(t *testing.T) {
	store := &checkoutStoreStub{}
	service := NewService(store, nil, nil, nil, nil, nil, nil)

	_, err := service.Create(t.Context(), uuid.New(), CreateParams{})
	if !errors.Is(err, ErrInvalidCheckout) {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalidCheckout)
	}
	if store.creates != 0 {
		t.Fatal("invalid checkout reached repository")
	}
}

func TestCreateWalletTopupUsesWalletCurrency(t *testing.T) {
	store := &checkoutStoreStub{}
	walletID := uuid.New()
	walletsService := &walletServiceStub{wallet: wallets.Wallet{
		ID:       walletID,
		Currency: "GHS",
		Status:   wallets.StatusActive,
	}}
	service := NewService(store, walletsService, nil, nil, nil, nil, nil)

	result, err := service.Create(t.Context(), uuid.New(), CreateParams{
		Type:        TypeWalletTopup,
		WalletID:    &walletID,
		AmountMinor: 5000,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if result.Currency != "GHS" || result.AmountMinor != 5000 {
		t.Fatalf("checkout terms = %s %d", result.Currency, result.AmountMinor)
	}
	if result.Provider != "" || result.PaymentMethod != "" {
		t.Fatal("new checkout must be provider-neutral")
	}
}

func TestCreateSubscriptionUsesCatalogPrice(t *testing.T) {
	store := &checkoutStoreStub{}
	priceID := uuid.New()
	planID := uuid.New()
	productID := uuid.New()
	amount := int64(2500)
	interval := catalog.BillingIntervalMonth
	now := time.Now().UTC().Add(-time.Minute)
	catalogService := &catalogServiceStub{
		price: catalog.Price{
			ID:              priceID,
			PlanID:          planID,
			PricingType:     catalog.PricingTypeRecurring,
			Currency:        "USD",
			AmountMinor:     &amount,
			BillingInterval: &interval,
			Active:          true,
			EffectiveFrom:   now,
		},
		plan:    catalog.Plan{ID: planID, ProductID: productID, Active: true},
		product: catalog.Product{ID: productID, Active: true},
	}
	subscriptionsService := &subscriptionServiceStub{currentErr: subscriptions.ErrSubscriptionNotFound}
	service := NewService(store, nil, catalogService, subscriptionsService, nil, nil, nil)

	result, err := service.Create(t.Context(), uuid.New(), CreateParams{
		Type:    TypeSubscription,
		PriceID: &priceID,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if result.AmountMinor != amount || result.Currency != "USD" {
		t.Fatalf("checkout terms = %s %d", result.Currency, result.AmountMinor)
	}
}

func TestConfirmSelectsProviderFromPaymentMethod(t *testing.T) {
	organizationID := uuid.New()
	checkoutID := uuid.New()
	store := &checkoutStoreStub{checkout: Checkout{
		ID:             checkoutID,
		OrganizationID: organizationID,
		Type:           TypeWalletTopup,
		Reference:      "checkout.test",
		AmountMinor:    5000,
		Currency:       "GHS",
		Status:         StatusPending,
		NextAction:     ActionWait,
	}}
	paymentsStore := &paymentStoreStub{}
	registry := commercialpayments.NewProviderRegistry(map[string]commercialpayments.Provider{
		"stripe": providerStub{session: commercialpayments.CheckoutSession{
			Provider:     "stripe",
			ProviderID:   "cs_123",
			Reference:    "checkout.test",
			ClientSecret: "secret",
			NextAction:   commercialpayments.NextActionWait,
			Status:       commercialpayments.StatusProcessing,
		}},
	})
	service := NewService(store, nil, nil, nil, paymentsStore, paymentEventsStub{}, registry)

	result, err := service.Confirm(t.Context(), organizationID, checkoutID, ConfirmInput{
		PaymentMethod: MethodCard,
		Email:         "billing@example.com",
	})
	if err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}
	if store.startedInput.Provider != ProviderStripe || paymentsStore.provider != "stripe" {
		t.Fatalf("provider selection = %q / %q", store.startedInput.Provider, paymentsStore.provider)
	}
	if result.Payment == nil || result.Session == nil {
		t.Fatal("confirmed checkout must include payment and provider session")
	}
}
