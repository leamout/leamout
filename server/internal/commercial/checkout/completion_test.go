package checkout

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

func TestCompletePaymentCreditsWalletAndCompletesCheckout(t *testing.T) {
	organizationID := uuid.New()
	checkoutID := uuid.New()
	walletID := uuid.New()
	paymentID := uuid.New()
	repository := &checkoutRepositoryStub{checkout: Checkout{
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
	service := newTestService(repository, walletsService, nil)
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
	if walletsService.postInput.Type != wallets.EntryTopup ||
		walletsService.postInput.SourceType != "checkout" ||
		walletsService.postInput.SourceID != checkoutID.String() ||
		walletsService.postInput.IdempotencyKey != "checkout:"+checkoutID.String() {
		t.Fatalf("wallet post = %+v", walletsService.postInput)
	}
	if repository.checkout.Status != StatusSucceeded {
		t.Fatalf("checkout status = %s, want %s", repository.checkout.Status, StatusSucceeded)
	}
}

func TestCompletePaymentTreatsDuplicateWalletCreditAsRetry(t *testing.T) {
	organizationID := uuid.New()
	checkoutID := uuid.New()
	walletID := uuid.New()
	repository := &checkoutRepositoryStub{checkout: Checkout{
		ID:             checkoutID,
		OrganizationID: organizationID,
		WalletID:       &walletID,
		Type:           TypeWalletTopup,
		Provider:       ProviderStripe,
		AmountMinor:    1000,
		Currency:       "USD",
		Status:         StatusProcessing,
	}}
	walletsService := &walletServiceStub{postErr: wallets.ErrDuplicateLedgerEntry}
	service := newTestService(repository, walletsService, nil)

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
	if repository.checkout.Status != StatusSucceeded {
		t.Fatalf("checkout status = %s, want succeeded", repository.checkout.Status)
	}
}

func TestCompletePaymentRejectsMismatchedSettlement(t *testing.T) {
	repository := &checkoutRepositoryStub{checkout: Checkout{
		ID:             uuid.New(),
		OrganizationID: uuid.New(),
		Type:           TypeWalletTopup,
		Provider:       ProviderStripe,
		AmountMinor:    5000,
		Currency:       "USD",
		Status:         StatusProcessing,
	}}
	service := newTestService(repository, &walletServiceStub{}, nil)

	err := service.CompletePayment(t.Context(), commercialpayments.Settlement{
		CheckoutID:     repository.checkout.ID,
		OrganizationID: repository.checkout.OrganizationID,
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
