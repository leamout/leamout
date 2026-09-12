package checkout

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/catalog"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/commercial/prepaid"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
)

func TestCompletePaymentCreditsWalletAndCompletesCheckout(t *testing.T) {
	organizationID := uuid.New()
	checkoutID := uuid.New()
	walletID := uuid.New()
	paymentID := uuid.New()
	store := &checkoutStoreStub{checkout: Checkout{
		ID:             checkoutID,
		OrganizationID: organizationID,
		WalletID:       &walletID,
		Type:           TypeWalletTopup,
		Provider:       ProviderStripe,
		PaymentMethod:  MethodCard,
		Reference:      "checkout.wallet",
		AmountMinor:    5000,
		Currency:       "GHS",
		Status:         StatusProcessing,
		NextAction:     ActionWait,
	}}
	walletsService := &walletServiceStub{}
	service := NewService(store, walletsService, nil, nil, nil)
	settledAt := time.Now().UTC()

	err := service.CompletePayment(t.Context(), commercialpayments.Settlement{
		Applied:        true,
		CheckoutID:     checkoutID,
		OrganizationID: organizationID,
		PaymentID:      paymentID,
		Provider:       "stripe",
		Status:         commercialpayments.StatusSucceeded,
		AmountMinor:    5000,
		Currency:       "GHS",
		SettledAt:      &settledAt,
	})
	if err != nil {
		t.Fatalf("CompletePayment() error = %v", err)
	}
	if walletsService.posts != 1 {
		t.Fatalf("wallet posts = %d, want 1", walletsService.posts)
	}
	if walletsService.postInput.Type != prepaid.EntryTopup ||
		walletsService.postInput.SourceType != "checkout" ||
		walletsService.postInput.SourceID != checkoutID.String() ||
		walletsService.postInput.IdempotencyKey != "checkout:"+checkoutID.String() {
		t.Fatalf("wallet post = %+v", walletsService.postInput)
	}
	if store.checkout.Status != StatusSucceeded {
		t.Fatalf("checkout status = %s, want %s", store.checkout.Status, StatusSucceeded)
	}
}

func TestCompletePaymentTreatsDuplicateWalletCreditAsRetry(t *testing.T) {
	organizationID := uuid.New()
	checkoutID := uuid.New()
	walletID := uuid.New()
	store := &checkoutStoreStub{checkout: Checkout{
		ID:             checkoutID,
		OrganizationID: organizationID,
		WalletID:       &walletID,
		Type:           TypeWalletTopup,
		Provider:       ProviderStripe,
		AmountMinor:    1000,
		Currency:       "USD",
		Status:         StatusProcessing,
	}}
	walletsService := &walletServiceStub{postErr: prepaid.ErrDuplicateLedgerEntry}
	service := NewService(store, walletsService, nil, nil, nil)

	err := service.CompletePayment(t.Context(), commercialpayments.Settlement{
		CheckoutID:     checkoutID,
		OrganizationID: organizationID,
		PaymentID:      uuid.New(),
		Provider:       "stripe",
		Status:         commercialpayments.StatusSucceeded,
		AmountMinor:    1000,
		Currency:       "USD",
	})
	if err != nil {
		t.Fatalf("CompletePayment() retry error = %v", err)
	}
	if store.checkout.Status != StatusSucceeded {
		t.Fatalf("checkout status = %s, want succeeded", store.checkout.Status)
	}
}

func TestCompletePaymentActivatesSubscription(t *testing.T) {
	organizationID := uuid.New()
	checkoutID := uuid.New()
	priceID := uuid.New()
	interval := catalog.BillingIntervalMonth
	store := &checkoutStoreStub{checkout: Checkout{
		ID:             checkoutID,
		OrganizationID: organizationID,
		PriceID:        &priceID,
		Type:           TypeSubscription,
		Provider:       ProviderStripe,
		PaymentMethod:  MethodCard,
		AmountMinor:    2500,
		Currency:       "USD",
		Status:         StatusProcessing,
	}}
	catalogService := &catalogServiceStub{price: catalog.Price{
		ID:              priceID,
		PricingType:     catalog.PricingTypeRecurring,
		BillingInterval: &interval,
	}}
	subscriptionsService := &subscriptionServiceStub{}
	service := NewService(store, nil, catalogService, subscriptionsService, nil)
	settledAt := time.Date(2026, 9, 11, 12, 0, 0, 0, time.UTC)

	err := service.CompletePayment(t.Context(), commercialpayments.Settlement{
		CheckoutID:     checkoutID,
		OrganizationID: organizationID,
		PaymentID:      uuid.New(),
		Provider:       "stripe",
		Status:         commercialpayments.StatusSucceeded,
		AmountMinor:    2500,
		Currency:       "USD",
		SettledAt:      &settledAt,
	})
	if err != nil {
		t.Fatalf("CompletePayment() error = %v", err)
	}
	if subscriptionsService.creates != 1 {
		t.Fatalf("subscription creates = %d, want 1", subscriptionsService.creates)
	}
	if subscriptionsService.created.Status == nil || *subscriptionsService.created.Status != subscriptions.StatusActive {
		t.Fatalf("subscription status = %v", subscriptionsService.created.Status)
	}
	if subscriptionsService.created.RenewsAt == nil || !subscriptionsService.created.RenewsAt.Equal(settledAt.AddDate(0, 1, 0)) {
		t.Fatalf("renews_at = %v", subscriptionsService.created.RenewsAt)
	}
}

func TestCompletePaymentRejectsMismatchedSettlement(t *testing.T) {
	store := &checkoutStoreStub{checkout: Checkout{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		Type:           TypeWalletTopup,
		Provider:       ProviderStripe,
		AmountMinor:    5000,
		Currency:       "USD",
		Status:         StatusProcessing,
	}}
	service := NewService(store, &walletServiceStub{}, nil, nil, nil)

	err := service.CompletePayment(t.Context(), commercialpayments.Settlement{
		CheckoutID:     store.checkout.ID,
		OrganizationID: store.checkout.OrganizationID,
		PaymentID:      uuid.New(),
		Provider:       "stripe",
		Status:         commercialpayments.StatusSucceeded,
		AmountMinor:    4000,
		Currency:       "USD",
	})
	if !errors.Is(err, ErrPaymentMismatch) {
		t.Fatalf("CompletePayment() error = %v, want %v", err, ErrPaymentMismatch)
	}
}
