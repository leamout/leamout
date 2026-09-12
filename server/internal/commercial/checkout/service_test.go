package checkout

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

type checkoutRepositoryStub struct {
	checkout     Checkout
	createdInput CreateInput
	startedInput StartPayment
	transition   Transition
	creates      int
	starts       int
	transitions  int
}

func (s *checkoutRepositoryStub) Create(_ context.Context, organizationID uuid.UUID, input CreateInput) (Checkout, error) {
	s.creates++
	s.createdInput = input
	result := s.checkout
	if result.ID == uuid.Nil {
		result.ID = uuid.New()
	}
	result.OrganizationID = organizationID
	result.WalletID = input.WalletID
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

func (s *checkoutRepositoryStub) StartPayment(_ context.Context, _, _ uuid.UUID, input StartPayment) (Checkout, error) {
	s.starts++
	s.startedInput = input
	s.checkout.Provider = input.Provider
	s.checkout.PaymentMethod = input.PaymentMethod
	s.checkout.Status = StatusProcessing
	s.checkout.NextAction = ActionWait
	return s.checkout, nil
}

func (s *checkoutRepositoryStub) Get(context.Context, uuid.UUID, uuid.UUID) (Checkout, error) {
	return s.checkout, nil
}

func (s *checkoutRepositoryStub) GetByReference(context.Context, string) (Checkout, error) {
	return s.checkout, nil
}

func (s *checkoutRepositoryStub) Transition(_ context.Context, _, _ uuid.UUID, input Transition) (Checkout, error) {
	s.transitions++
	s.transition = input
	s.checkout.Status = input.Status
	s.checkout.NextAction = input.NextAction
	s.checkout.ProviderMessage = input.ProviderMessage
	s.checkout.CompletedAt = input.CompletedAt
	return s.checkout, nil
}

func (s *checkoutRepositoryStub) ClaimRefresh(context.Context, uuid.UUID, uuid.UUID, time.Time) (Checkout, error) {
	return Checkout{}, ErrCheckoutNotFound
}

func (s *checkoutRepositoryStub) Expire(context.Context) ([]Checkout, error) { return nil, nil }

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

type paymentServiceStub struct {
	available  bool
	payment    commercialpayments.Payment
	startInput commercialpayments.StartInput
	start      commercialpayments.StartResult
	startErr   error
}

func (s *paymentServiceStub) ProviderAvailable(string) bool { return s.available }

func (s *paymentServiceStub) GetByCheckout(context.Context, uuid.UUID, uuid.UUID) (commercialpayments.Payment, error) {
	if s.payment.ID == uuid.Nil {
		return commercialpayments.Payment{}, commercialpayments.ErrPaymentNotFound
	}
	return s.payment, nil
}

func (s *paymentServiceStub) Start(
	_ context.Context,
	_ uuid.UUID,
	input commercialpayments.StartInput,
) (commercialpayments.StartResult, error) {
	s.startInput = input
	return s.start, s.startErr
}

func (s *paymentServiceStub) Continue(context.Context, commercialpayments.ContinueInput) (commercialpayments.CheckoutSession, error) {
	return commercialpayments.CheckoutSession{}, nil
}

func (s *paymentServiceStub) Refresh(context.Context, string, string) (commercialpayments.Settlement, bool, error) {
	return commercialpayments.Settlement{}, false, nil
}

func TestServiceRejectsInvalidCheckoutBeforePersistence(t *testing.T) {
	repository := &checkoutRepositoryStub{}
	service := newTestService(repository, nil, nil)

	_, err := service.Create(t.Context(), uuid.New(), CreateParams{})
	if !errors.Is(err, ErrInvalidCheckout) {
		t.Fatalf("Create() error = %v, want %v", err, ErrInvalidCheckout)
	}
	if repository.creates != 0 {
		t.Fatal("invalid checkout reached repository")
	}
}

func TestCreateWalletTopupUsesWalletCurrency(t *testing.T) {
	repository := &checkoutRepositoryStub{}
	walletID := uuid.New()
	walletsService := &walletServiceStub{wallet: wallets.Wallet{
		ID:       walletID,
		Currency: "GHS",
		Status:   wallets.StatusActive,
	}}
	service := newTestService(repository, walletsService, nil)

	result, err := service.Create(t.Context(), uuid.New(), CreateParams{
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

func TestConfirmDelegatesCollectionToPayments(t *testing.T) {
	organizationID := uuid.New()
	checkoutID := uuid.New()
	walletID := uuid.New()
	repository := &checkoutRepositoryStub{checkout: Checkout{
		ID:             checkoutID,
		OrganizationID: organizationID,
		WalletID:       &walletID,
		Reference:      "checkout.test",
		AmountMinor:    5000,
		Currency:       "GHS",
		Status:         StatusPending,
		NextAction:     ActionWait,
	}}
	paymentID := uuid.New()
	paymentsService := &paymentServiceStub{
		available: true,
		start: commercialpayments.StartResult{
			Payment: commercialpayments.Payment{ID: paymentID, CheckoutID: checkoutID, Provider: "stripe", Status: commercialpayments.StatusProcessing},
			Session: commercialpayments.CheckoutSession{
				Provider:     "stripe",
				ProviderID:   "cs_123",
				Reference:    "checkout.test",
				ClientSecret: "secret",
				NextAction:   commercialpayments.NextActionWait,
				Status:       commercialpayments.StatusProcessing,
			},
		},
	}
	service := newTestService(repository, nil, paymentsService)

	result, err := service.Confirm(t.Context(), organizationID, checkoutID, ConfirmInput{
		PaymentMethod: MethodCard,
		Email:         "billing@example.com",
	})
	if err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}
	if repository.startedInput.Provider != ProviderStripe || paymentsService.startInput.Provider != "stripe" {
		t.Fatalf("provider selection = %q / %q", repository.startedInput.Provider, paymentsService.startInput.Provider)
	}
	if result.Payment == nil || result.Session == nil || result.Payment.ID != paymentID {
		t.Fatal("confirmed checkout must include payment and provider session")
	}
}

func newTestService(
	repo *checkoutRepositoryStub,
	wallet *walletServiceStub,
	payments *paymentServiceStub,
) *Service {
	service := &Service{now: time.Now}
	if repo != nil {
		service.checkouts = checkoutOperations{
			create:         repo.Create,
			startPayment:   repo.StartPayment,
			get:            repo.Get,
			getByReference: repo.GetByReference,
			transition:     repo.Transition,
			claimRefresh:   repo.ClaimRefresh,
			expire:         repo.Expire,
		}
	}
	if wallet != nil {
		service.wallets = walletOperations{get: wallet.Get, post: wallet.Post}
	}
	if payments != nil {
		service.payments = paymentOperations{
			providerAvailable: payments.ProviderAvailable,
			getByCheckout:     payments.GetByCheckout,
			start:             payments.Start,
			continuePayment:   payments.Continue,
			refresh:           payments.Refresh,
		}
	}
	return service
}
