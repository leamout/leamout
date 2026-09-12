package checkout

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

type checkoutOperations struct {
	create         func(context.Context, uuid.UUID, CreateInput) (Checkout, error)
	startPayment   func(context.Context, uuid.UUID, uuid.UUID, StartPayment) (Checkout, error)
	get            func(context.Context, uuid.UUID, uuid.UUID) (Checkout, error)
	getByReference func(context.Context, string) (Checkout, error)
	transition     func(context.Context, uuid.UUID, uuid.UUID, Transition) (Checkout, error)
	claimRefresh   func(context.Context, uuid.UUID, uuid.UUID, time.Time) (Checkout, error)
	expire         func(context.Context) ([]Checkout, error)
}

type walletOperations struct {
	get  func(context.Context, uuid.UUID, uuid.UUID) (wallets.Wallet, error)
	post func(context.Context, uuid.UUID, uuid.UUID, wallets.PostEntryInput) (wallets.LedgerEntry, error)
}

type paymentOperations struct {
	providerAvailable func(string) bool
	getByCheckout     func(context.Context, uuid.UUID, uuid.UUID) (commercialpayments.Payment, error)
	start             func(context.Context, uuid.UUID, commercialpayments.StartInput) (commercialpayments.StartResult, error)
	continuePayment   func(context.Context, commercialpayments.ContinueInput) (commercialpayments.CheckoutSession, error)
	refresh           func(context.Context, string, string) (commercialpayments.Settlement, bool, error)
}

type Service struct {
	checkouts checkoutOperations
	wallets   walletOperations
	payments  paymentOperations
	now       func() time.Time
}

type ConfirmInput struct {
	PaymentMethod PaymentMethod
	Email         string
	CallbackURL   string
	MobileMoney   *commercialpayments.MobileMoney
}

type ContinueInput struct {
	Action commercialpayments.NextAction
	Value  string
}

type Result struct {
	Checkout Checkout
	Payment  *commercialpayments.Payment
	Session  *commercialpayments.CheckoutSession
}

func NewService(repo *Repository, walletService *wallets.Service, paymentService *commercialpayments.Service) *Service {
	if repo == nil || walletService == nil || paymentService == nil {
		panic("checkout: repository, wallet service, and payment service are required")
	}
	return &Service{
		checkouts: checkoutOperations{
			create: repo.Create, startPayment: repo.StartPayment, get: repo.Get,
			getByReference: repo.GetByReference, transition: repo.Transition,
			claimRefresh: repo.ClaimRefresh, expire: repo.Expire,
		},
		wallets: walletOperations{get: walletService.Get, post: walletService.Post},
		payments: paymentOperations{
			providerAvailable: paymentService.ProviderAvailable,
			getByCheckout:     paymentService.GetByCheckout, start: paymentService.Start,
			continuePayment: paymentService.Continue, refresh: paymentService.Refresh,
		},
		now: time.Now,
	}
}

func (s *Service) Create(ctx context.Context, organizationID uuid.UUID, params CreateParams) (Checkout, error) {
	if organizationID == uuid.Nil || validateIntent(params) != nil {
		return Checkout{}, ErrInvalidCheckout
	}
	wallet, err := s.wallets.get(ctx, organizationID, *params.WalletID)
	if err != nil {
		return Checkout{}, err
	}
	if wallet.Status != wallets.StatusActive {
		return Checkout{}, ErrInvalidCheckout
	}
	now := s.now().UTC()
	input := CreateInput{
		WalletID: params.WalletID, Type: TypeWalletTopup,
		Reference: "checkout." + uuid.NewString(), AmountMinor: params.AmountMinor,
		Currency: wallet.Currency, ExpiresAt: now.Add(30 * time.Minute), Metadata: params.Metadata,
	}
	if validateCreate(input, now) != nil {
		return Checkout{}, ErrInvalidCheckout
	}
	return s.checkouts.create(ctx, organizationID, input)
}

func (s *Service) Get(ctx context.Context, organizationID, checkoutID uuid.UUID) (Result, error) {
	checkoutRecord, err := s.checkouts.get(ctx, organizationID, checkoutID)
	if err != nil {
		return Result{}, err
	}
	payment, err := s.payments.getByCheckout(ctx, organizationID, checkoutID)
	if errors.Is(err, commercialpayments.ErrPaymentNotFound) {
		return Result{Checkout: checkoutRecord}, nil
	}
	if err != nil {
		return Result{}, err
	}
	if checkoutRecord.Provider == ProviderPaystack && checkoutRecord.Status == StatusProcessing {
		claimed, claimErr := s.checkouts.claimRefresh(ctx, organizationID, checkoutID, s.now().UTC().Add(-10*time.Second))
		if claimErr == nil {
			settlement, refreshed, refreshErr := s.payments.refresh(ctx, string(claimed.Provider), claimed.Reference)
			if refreshErr != nil {
				return Result{}, refreshErr
			}
			if refreshed && settlement.CheckoutID != uuid.Nil {
				if err := s.CompletePayment(ctx, settlement); err != nil {
					return Result{}, err
				}
				checkoutRecord, err = s.checkouts.get(ctx, organizationID, checkoutID)
				if err != nil {
					return Result{}, err
				}
				payment, err = s.payments.getByCheckout(ctx, organizationID, checkoutID)
				if err != nil {
					return Result{}, err
				}
			}
		} else if !errors.Is(claimErr, ErrCheckoutNotFound) {
			return Result{}, claimErr
		}
	}
	return Result{Checkout: checkoutRecord, Payment: &payment}, nil
}

func (s *Service) Confirm(ctx context.Context, organizationID, checkoutID uuid.UUID, input ConfirmInput) (Result, error) {
	if organizationID == uuid.Nil || checkoutID == uuid.Nil || validateConfirm(input) != nil {
		return Result{}, ErrInvalidCheckout
	}
	checkoutRecord, err := s.checkouts.get(ctx, organizationID, checkoutID)
	if err != nil {
		return Result{}, err
	}
	if checkoutRecord.Status != StatusPending || checkoutRecord.Type != TypeWalletTopup {
		return Result{}, ErrInvalidTransition
	}
	providerName, ok := providerForMethod(input.PaymentMethod)
	if !ok {
		return Result{}, ErrInvalidCheckout
	}
	if providerName == ProviderPaystack && checkoutRecord.Currency != "GHS" {
		return Result{}, ErrInvalidCheckout
	}
	if !s.payments.providerAvailable(string(providerName)) {
		return Result{}, ErrProviderUnavailable
	}
	checkoutRecord, err = s.checkouts.startPayment(ctx, organizationID, checkoutID, StartPayment{Provider: providerName, PaymentMethod: input.PaymentMethod})
	if err != nil {
		return Result{}, err
	}
	metadata := map[string]string{
		"organization_id": organizationID.String(), "checkout_id": checkoutRecord.ID.String(),
		"checkout_type": string(checkoutRecord.Type), "wallet_id": checkoutRecord.WalletID.String(),
	}
	started, err := s.payments.start(ctx, organizationID, commercialpayments.StartInput{
		CheckoutID: checkoutRecord.ID, Provider: string(providerName), Reference: checkoutRecord.Reference,
		AmountMinor: checkoutRecord.AmountMinor, Currency: checkoutRecord.Currency,
		Email: input.Email, CallbackURL: input.CallbackURL, Metadata: metadata, MobileMoney: input.MobileMoney,
	})
	if err != nil {
		s.failProcessingCheckout(ctx, organizationID, checkoutRecord)
		return Result{}, err
	}
	message := strings.TrimSpace(started.Session.Message)
	var providerMessage *string
	if message != "" {
		providerMessage = &message
	}
	nextAction := NextAction(started.Session.NextAction)
	if nextAction == "" {
		nextAction = ActionWait
	}
	checkoutRecord, err = s.checkouts.transition(ctx, organizationID, checkoutID, Transition{
		Expected: StatusProcessing, Status: StatusProcessing, NextAction: nextAction, ProviderMessage: providerMessage,
	})
	if err != nil {
		return Result{}, err
	}
	return Result{Checkout: checkoutRecord, Payment: &started.Payment, Session: &started.Session}, nil
}

func (s *Service) Continue(ctx context.Context, organizationID, checkoutID uuid.UUID, input ContinueInput) (Result, error) {
	details, err := s.Get(ctx, organizationID, checkoutID)
	if err != nil {
		return Result{}, err
	}
	if details.Payment == nil || details.Checkout.Provider != ProviderPaystack || details.Checkout.Status != StatusProcessing {
		return Result{}, ErrInvalidCheckout
	}
	session, err := s.payments.continuePayment(ctx, commercialpayments.ContinueInput{
		Provider: string(details.Checkout.Provider), Reference: details.Checkout.Reference, Action: input.Action, Value: input.Value,
	})
	if err != nil {
		return Result{}, err
	}
	message := strings.TrimSpace(session.Message)
	var providerMessage *string
	if message != "" {
		providerMessage = &message
	}
	nextAction := NextAction(session.NextAction)
	if nextAction == "" {
		nextAction = ActionWait
	}
	checkoutRecord, err := s.checkouts.transition(ctx, organizationID, checkoutID, Transition{
		Expected: StatusProcessing, Status: StatusProcessing, NextAction: nextAction, ProviderMessage: providerMessage,
	})
	if err != nil {
		return Result{}, err
	}
	return Result{Checkout: checkoutRecord, Payment: details.Payment, Session: &session}, nil
}

func (s *Service) GetByReference(ctx context.Context, reference string) (Checkout, error) {
	return s.checkouts.getByReference(ctx, reference)
}

func (s *Service) Transition(ctx context.Context, organizationID, id uuid.UUID, transition Transition) (Checkout, error) {
	if validateTransition(transition) != nil {
		return Checkout{}, ErrInvalidTransition
	}
	return s.checkouts.transition(ctx, organizationID, id, transition)
}

func (s *Service) Expire(ctx context.Context) ([]Checkout, error) { return s.checkouts.expire(ctx) }

func (s *Service) failProcessingCheckout(ctx context.Context, organizationID uuid.UUID, checkoutRecord Checkout) {
	completedAt := s.now().UTC()
	_, _ = s.checkouts.transition(ctx, organizationID, checkoutRecord.ID, Transition{
		Expected: StatusProcessing, Status: StatusFailed, NextAction: ActionNone, CompletedAt: &completedAt,
	})
}
