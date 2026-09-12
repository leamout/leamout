package checkout

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/catalog"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/commercial/subscriptions"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

type store interface {
	Create(context.Context, uuid.UUID, CreateInput) (Checkout, error)
	StartPayment(context.Context, uuid.UUID, uuid.UUID, StartPayment) (Checkout, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (Checkout, error)
	GetByReference(context.Context, string) (Checkout, error)
	Transition(context.Context, uuid.UUID, uuid.UUID, Transition) (Checkout, error)
	ClaimRefresh(context.Context, uuid.UUID, uuid.UUID, time.Time) (Checkout, error)
	Expire(context.Context) ([]Checkout, error)
}

type walletService interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (wallets.Wallet, error)
	Post(context.Context, uuid.UUID, uuid.UUID, wallets.PostEntryInput) (wallets.LedgerEntry, error)
}

type catalogService interface {
	GetPrice(context.Context, uuid.UUID) (catalog.Price, error)
	GetPlan(context.Context, uuid.UUID) (catalog.Plan, error)
	GetProduct(context.Context, uuid.UUID) (catalog.Product, error)
}

type subscriptionService interface {
	Create(context.Context, uuid.UUID, subscriptions.CreateInput) (subscriptions.Subscription, error)
	Current(context.Context, uuid.UUID) (subscriptions.Subscription, error)
}

type paymentService interface {
	ProviderAvailable(string) bool
	GetByCheckout(context.Context, uuid.UUID, uuid.UUID) (commercialpayments.Payment, error)
	Start(context.Context, uuid.UUID, commercialpayments.StartInput) (commercialpayments.StartResult, error)
	Continue(context.Context, commercialpayments.ContinueInput) (commercialpayments.CheckoutSession, error)
	Refresh(context.Context, string, string) (commercialpayments.Settlement, bool, error)
}

type Service struct {
	repo          store
	wallets       walletService
	catalog       catalogService
	subscriptions subscriptionService
	payments      paymentService
	now           func() time.Time
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

func NewService(
	repo store,
	wallets walletService,
	catalog catalogService,
	subscriptions subscriptionService,
	payments paymentService,
) *Service {
	return &Service{
		repo:          repo,
		wallets:       wallets,
		catalog:       catalog,
		subscriptions: subscriptions,
		payments:      payments,
		now:           time.Now,
	}
}

func (s *Service) Create(ctx context.Context, organizationID uuid.UUID, params CreateParams) (Checkout, error) {
	if organizationID == uuid.Nil || validateIntent(params) != nil {
		return Checkout{}, ErrInvalidCheckout
	}

	now := s.now().UTC()
	input := CreateInput{
		WalletID:  params.WalletID,
		PriceID:   params.PriceID,
		Type:      params.Type,
		Reference: "checkout." + uuid.NewString(),
		ExpiresAt: now.Add(30 * time.Minute),
		Metadata:  params.Metadata,
	}

	switch params.Type {
	case TypeWalletTopup:
		wallet, err := s.wallets.Get(ctx, organizationID, *params.WalletID)
		if err != nil {
			return Checkout{}, err
		}
		if wallet.Status != wallets.StatusActive {
			return Checkout{}, ErrInvalidCheckout
		}
		input.AmountMinor = params.AmountMinor
		input.Currency = wallet.Currency

	case TypeSubscription:
		if _, err := s.subscriptions.Current(ctx, organizationID); err == nil {
			return Checkout{}, subscriptions.ErrCurrentSubscriptionExists
		} else if !errors.Is(err, subscriptions.ErrSubscriptionNotFound) {
			return Checkout{}, err
		}

		price, err := s.catalog.GetPrice(ctx, *params.PriceID)
		if err != nil {
			return Checkout{}, err
		}
		if price.PricingType != catalog.PricingTypeRecurring ||
			price.AmountMinor == nil || *price.AmountMinor <= 0 ||
			price.BillingInterval == nil || !price.EffectiveAt(now) {
			return Checkout{}, ErrInvalidCheckout
		}
		plan, err := s.catalog.GetPlan(ctx, price.PlanID)
		if err != nil {
			return Checkout{}, err
		}
		if !plan.Active {
			return Checkout{}, ErrInvalidCheckout
		}
		product, err := s.catalog.GetProduct(ctx, plan.ProductID)
		if err != nil {
			return Checkout{}, err
		}
		if !product.Active {
			return Checkout{}, ErrInvalidCheckout
		}
		input.AmountMinor = *price.AmountMinor
		input.Currency = price.Currency
	}

	if validateCreate(input, now) != nil {
		return Checkout{}, ErrInvalidCheckout
	}
	return s.repo.Create(ctx, organizationID, input)
}

func (s *Service) Get(ctx context.Context, organizationID, checkoutID uuid.UUID) (Result, error) {
	checkoutRecord, err := s.repo.Get(ctx, organizationID, checkoutID)
	if err != nil {
		return Result{}, err
	}
	if s.payments == nil {
		return Result{Checkout: checkoutRecord}, nil
	}

	payment, err := s.payments.GetByCheckout(ctx, organizationID, checkoutID)
	if errors.Is(err, commercialpayments.ErrPaymentNotFound) {
		return Result{Checkout: checkoutRecord}, nil
	}
	if err != nil {
		return Result{}, err
	}

	if checkoutRecord.Provider == ProviderPaystack && checkoutRecord.Status == StatusProcessing {
		claimed, claimErr := s.repo.ClaimRefresh(
			ctx,
			organizationID,
			checkoutID,
			s.now().UTC().Add(-10*time.Second),
		)
		if claimErr == nil {
			settlement, refreshed, refreshErr := s.payments.Refresh(ctx, string(claimed.Provider), claimed.Reference)
			if refreshErr != nil {
				return Result{}, refreshErr
			}
			if refreshed && settlement.CheckoutID != uuid.Nil {
				if refreshErr = s.CompletePayment(ctx, settlement); refreshErr != nil {
					return Result{}, refreshErr
				}
				checkoutRecord, err = s.repo.Get(ctx, organizationID, checkoutID)
				if err != nil {
					return Result{}, err
				}
				payment, err = s.payments.GetByCheckout(ctx, organizationID, checkoutID)
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

func (s *Service) Confirm(
	ctx context.Context,
	organizationID, checkoutID uuid.UUID,
	input ConfirmInput,
) (Result, error) {
	if organizationID == uuid.Nil || checkoutID == uuid.Nil || validateConfirm(input) != nil || s.payments == nil {
		return Result{}, ErrInvalidCheckout
	}

	checkoutRecord, err := s.repo.Get(ctx, organizationID, checkoutID)
	if err != nil {
		return Result{}, err
	}
	if checkoutRecord.Status != StatusPending {
		return Result{}, ErrInvalidTransition
	}

	providerName, ok := providerForMethod(input.PaymentMethod)
	if !ok {
		return Result{}, ErrInvalidCheckout
	}
	if providerName == ProviderPaystack && checkoutRecord.Currency != "GHS" {
		return Result{}, ErrInvalidCheckout
	}
	if !s.payments.ProviderAvailable(string(providerName)) {
		return Result{}, ErrProviderUnavailable
	}

	checkoutRecord, err = s.repo.StartPayment(ctx, organizationID, checkoutID, StartPayment{
		Provider:      providerName,
		PaymentMethod: input.PaymentMethod,
	})
	if err != nil {
		return Result{}, err
	}

	metadata := map[string]string{
		"organization_id": organizationID.String(),
		"checkout_id":     checkoutRecord.ID.String(),
		"checkout_type":   string(checkoutRecord.Type),
	}
	if checkoutRecord.WalletID != nil {
		metadata["wallet_id"] = checkoutRecord.WalletID.String()
	}
	if checkoutRecord.PriceID != nil {
		metadata["price_id"] = checkoutRecord.PriceID.String()
	}

	started, err := s.payments.Start(ctx, organizationID, commercialpayments.StartInput{
		CheckoutID:  checkoutRecord.ID,
		Provider:    string(providerName),
		Reference:   checkoutRecord.Reference,
		AmountMinor: checkoutRecord.AmountMinor,
		Currency:    checkoutRecord.Currency,
		Email:       input.Email,
		CallbackURL: input.CallbackURL,
		Metadata:    metadata,
		MobileMoney: input.MobileMoney,
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

	checkoutRecord, err = s.repo.Transition(
		ctx,
		organizationID,
		checkoutRecord.ID,
		Transition{
			Expected:        StatusProcessing,
			Status:          StatusProcessing,
			NextAction:      nextAction,
			ProviderMessage: providerMessage,
		},
	)
	if err != nil {
		return Result{}, err
	}

	return Result{Checkout: checkoutRecord, Payment: &started.Payment, Session: &started.Session}, nil
}

func (s *Service) Continue(
	ctx context.Context,
	organizationID, checkoutID uuid.UUID,
	input ContinueInput,
) (Result, error) {
	details, err := s.Get(ctx, organizationID, checkoutID)
	if err != nil {
		return Result{}, err
	}
	if details.Payment == nil || s.payments == nil ||
		details.Checkout.Provider != ProviderPaystack || details.Checkout.Status != StatusProcessing {
		return Result{}, ErrInvalidCheckout
	}

	session, err := s.payments.Continue(ctx, commercialpayments.ContinueInput{
		Provider:  string(details.Checkout.Provider),
		Reference: details.Checkout.Reference,
		Action:    input.Action,
		Value:     input.Value,
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

	checkoutRecord, err := s.repo.Transition(
		ctx,
		organizationID,
		checkoutID,
		Transition{
			Expected:        StatusProcessing,
			Status:          StatusProcessing,
			NextAction:      nextAction,
			ProviderMessage: providerMessage,
		},
	)
	if err != nil {
		return Result{}, err
	}

	return Result{Checkout: checkoutRecord, Payment: details.Payment, Session: &session}, nil
}

func (s *Service) GetByReference(ctx context.Context, reference string) (Checkout, error) {
	return s.repo.GetByReference(ctx, reference)
}

func (s *Service) Transition(
	ctx context.Context,
	organizationID, id uuid.UUID,
	transition Transition,
) (Checkout, error) {
	if validateTransition(transition) != nil {
		return Checkout{}, ErrInvalidTransition
	}
	return s.repo.Transition(ctx, organizationID, id, transition)
}

func (s *Service) Expire(ctx context.Context) ([]Checkout, error) {
	return s.repo.Expire(ctx)
}

func (s *Service) failProcessingCheckout(ctx context.Context, organizationID uuid.UUID, checkoutRecord Checkout) {
	completedAt := s.now().UTC()
	_, _ = s.repo.Transition(
		ctx,
		organizationID,
		checkoutRecord.ID,
		Transition{
			Expected:    StatusProcessing,
			Status:      StatusFailed,
			NextAction:  ActionNone,
			CompletedAt: &completedAt,
		},
	)
}
