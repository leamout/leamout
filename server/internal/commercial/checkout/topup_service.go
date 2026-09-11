package checkout

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/commercial/wallets"
)

type walletStore interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (wallets.Wallet, error)
}

type checkoutStore interface {
	Create(context.Context, uuid.UUID, CreateInput) (Checkout, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (Checkout, error)
	ClaimRefresh(context.Context, uuid.UUID, uuid.UUID, time.Time) (Checkout, error)
	Transition(context.Context, uuid.UUID, uuid.UUID, Transition) (Checkout, error)
}

type paymentStore interface {
	Create(context.Context, uuid.UUID, string, commercialpayments.CreateInput) (commercialpayments.Payment, error)
	GetByCheckout(context.Context, uuid.UUID, uuid.UUID) (commercialpayments.Payment, error)
	SetProviderID(context.Context, uuid.UUID, uuid.UUID, string, commercialpayments.Status) (commercialpayments.Payment, error)
	UpdateStatus(context.Context, uuid.UUID, uuid.UUID, commercialpayments.Status, *time.Time) (commercialpayments.Payment, error)
}

type paymentEventProcessor interface {
	ProcessProviderEvent(context.Context, commercialpayments.ProviderEvent) (commercialpayments.Settlement, error)
}

type TopupService struct {
	wallets       walletStore
	checkouts     checkoutStore
	payments      paymentStore
	paymentEvents paymentEventProcessor
	providers     *commercialpayments.ProviderRegistry
	now           func() time.Time
}

func NewTopupService(
	wallets walletStore,
	checkouts checkoutStore,
	payments paymentStore,
	paymentEvents paymentEventProcessor,
	providers *commercialpayments.ProviderRegistry,
) *TopupService {
	return &TopupService{
		wallets:       wallets,
		checkouts:     checkouts,
		payments:      payments,
		paymentEvents: paymentEvents,
		providers:     providers,
		now:           time.Now,
	}
}

func (s *TopupService) Create(
	ctx context.Context,
	organizationID uuid.UUID,
	walletID uuid.UUID,
	input TopupCreateInput,
) (TopupResult, error) {
	providerName := string(input.Provider)
	provider, ok := s.providers.Get(providerName)
	if !ok {
		return TopupResult{}, ErrProviderUnavailable
	}

	wallet, err := s.wallets.Get(ctx, organizationID, walletID)
	if err != nil {
		return TopupResult{}, err
	}
	if wallet.Status != wallets.StatusActive || input.AmountMinor <= 0 || strings.TrimSpace(input.Email) == "" {
		return TopupResult{}, ErrInvalidTopup
	}

	method := MethodCard
	if input.Provider == ProviderPaystack {
		method = MethodMobileMoney
		if wallet.Currency != "GHS" || input.MobileMoney == nil {
			return TopupResult{}, ErrInvalidTopup
		}
	} else if input.Provider != ProviderStripe || input.MobileMoney != nil {
		return TopupResult{}, ErrInvalidTopup
	}

	now := s.now().UTC()
	reference := "topup." + uuid.NewString()
	checkoutRecord, err := s.checkouts.Create(ctx, organizationID, CreateInput{
		WalletID:      &walletID,
		Type:          TypeWalletTopup,
		Provider:      input.Provider,
		PaymentMethod: method,
		Reference:     reference,
		AmountMinor:   input.AmountMinor,
		Currency:      wallet.Currency,
		ExpiresAt:     now.Add(30 * time.Minute),
	})
	if err != nil {
		return TopupResult{}, err
	}

	payment, err := s.payments.Create(ctx, organizationID, providerName, commercialpayments.CreateInput{
		CheckoutID:  checkoutRecord.ID,
		Status:      commercialpayments.StatusPending,
		AmountMinor: checkoutRecord.AmountMinor,
		Currency:    checkoutRecord.Currency,
	})
	if err != nil {
		return TopupResult{}, err
	}

	session, err := provider.CreateCheckout(ctx, commercialpayments.CheckoutRequest{
		Reference:   reference,
		AmountMinor: checkoutRecord.AmountMinor,
		Currency:    checkoutRecord.Currency,
		Email:       strings.TrimSpace(input.Email),
		CallbackURL: strings.TrimSpace(input.CallbackURL),
		Metadata: map[string]string{
			"organization_id": organizationID.String(),
			"wallet_id":       walletID.String(),
		},
		MobileMoney: input.MobileMoney,
	})
	if err != nil {
		s.failPendingCheckout(ctx, organizationID, checkoutRecord, payment)
		return TopupResult{}, err
	}
	if session.Provider != providerName || session.Reference != reference ||
		(input.Provider == ProviderStripe && session.ProviderID == "") {
		s.failPendingCheckout(ctx, organizationID, checkoutRecord, payment)
		return TopupResult{}, ErrPaymentMismatch
	}

	if session.ProviderID == "" {
		payment, err = s.payments.UpdateStatus(
			ctx,
			organizationID,
			payment.ID,
			commercialpayments.StatusProcessing,
			nil,
		)
	} else {
		payment, err = s.payments.SetProviderID(
			ctx,
			organizationID,
			payment.ID,
			session.ProviderID,
			commercialpayments.StatusProcessing,
		)
	}
	if err != nil {
		return TopupResult{}, err
	}

	message := strings.TrimSpace(session.Message)
	var providerMessage *string
	if message != "" {
		providerMessage = &message
	}

	checkoutRecord, err = s.checkouts.Transition(
		ctx,
		organizationID,
		checkoutRecord.ID,
		Transition{
			Expected:        StatusPending,
			Status:          StatusProcessing,
			NextAction:      NextAction(session.NextAction),
			ProviderMessage: providerMessage,
		},
	)
	if err != nil {
		return TopupResult{}, err
	}

	return TopupResult{
		Checkout: checkoutRecord,
		Payment:  payment,
		Session:  session,
	}, nil
}

func (s *TopupService) failPendingCheckout(
	ctx context.Context,
	organizationID uuid.UUID,
	checkoutRecord Checkout,
	payment commercialpayments.Payment,
) {
	completedAt := s.now().UTC()

	_, _ = s.payments.UpdateStatus(
		ctx,
		organizationID,
		payment.ID,
		commercialpayments.StatusFailed,
		nil,
	)
	_, _ = s.checkouts.Transition(
		ctx,
		organizationID,
		checkoutRecord.ID,
		Transition{
			Expected:    StatusPending,
			Status:      StatusFailed,
			NextAction:  ActionNone,
			CompletedAt: &completedAt,
		},
	)
}

func (s *TopupService) Get(
	ctx context.Context,
	organizationID uuid.UUID,
	checkoutID uuid.UUID,
) (TopupDetails, error) {
	checkoutRecord, err := s.checkouts.Get(ctx, organizationID, checkoutID)
	if err != nil {
		return TopupDetails{}, err
	}

	payment, err := s.payments.GetByCheckout(ctx, organizationID, checkoutID)
	if err != nil {
		return TopupDetails{}, err
	}

	if checkoutRecord.Provider == ProviderPaystack && checkoutRecord.Status == StatusProcessing {
		claimed, claimErr := s.checkouts.ClaimRefresh(
			ctx,
			organizationID,
			checkoutID,
			s.now().UTC().Add(-10*time.Second),
		)
		if claimErr == nil {
			if refreshed, ok := s.refreshPaystack(ctx, claimed); ok {
				if _, err = s.paymentEvents.ProcessProviderEvent(ctx, refreshed); err != nil {
					return TopupDetails{}, err
				}

				checkoutRecord, err = s.checkouts.Get(ctx, organizationID, checkoutID)
				if err != nil {
					return TopupDetails{}, err
				}

				payment, err = s.payments.GetByCheckout(ctx, organizationID, checkoutID)
				if err != nil {
					return TopupDetails{}, err
				}
			}
		} else if !errors.Is(claimErr, ErrCheckoutNotFound) {
			return TopupDetails{}, claimErr
		}
	}

	return TopupDetails{
		Checkout: checkoutRecord,
		Payment:  payment,
	}, nil
}

func (s *TopupService) refreshPaystack(
	ctx context.Context,
	checkoutRecord Checkout,
) (commercialpayments.ProviderEvent, bool) {
	provider, ok := s.providers.Get("paystack")
	if !ok {
		return commercialpayments.ProviderEvent{}, false
	}

	payment, err := provider.GetPayment(ctx, checkoutRecord.Reference)
	if err != nil || (payment.Status != commercialpayments.StatusSucceeded && payment.Status != commercialpayments.StatusFailed) {
		return commercialpayments.ProviderEvent{}, false
	}

	raw, _ := json.Marshal(map[string]string{
		"source":    "charge_lookup",
		"reference": checkoutRecord.Reference,
		"status":    string(payment.Status),
	})

	eventType := "charge.failed"
	if payment.Status == commercialpayments.StatusSucceeded {
		eventType = "charge.success"
	}

	return commercialpayments.ProviderEvent{
		Provider:        "paystack",
		ProviderEventID: "lookup:" + checkoutRecord.Reference + ":" + string(payment.Status),
		Type:            eventType,
		Payment:         payment,
		Raw:             raw,
	}, true
}

func (s *TopupService) Continue(
	ctx context.Context,
	organizationID uuid.UUID,
	checkoutID uuid.UUID,
	input TopupContinueInput,
) (TopupResult, error) {
	details, err := s.Get(ctx, organizationID, checkoutID)
	if err != nil {
		return TopupResult{}, err
	}

	provider, ok := s.providers.Get(string(details.Checkout.Provider))
	if !ok {
		return TopupResult{}, ErrProviderUnavailable
	}

	continuation, ok := provider.(commercialpayments.ContinuationProvider)
	if !ok || details.Checkout.Provider != ProviderPaystack ||
		details.Checkout.Status != StatusProcessing {
		return TopupResult{}, ErrInvalidTopup
	}

	session, err := continuation.ContinueCheckout(ctx, commercialpayments.ContinueCheckoutRequest{
		Reference: details.Checkout.Reference,
		Action:    input.Action,
		Value:     input.Value,
	})
	if err != nil {
		return TopupResult{}, err
	}

	message := strings.TrimSpace(session.Message)
	var providerMessage *string
	if message != "" {
		providerMessage = &message
	}

	checkoutRecord, err := s.checkouts.Transition(
		ctx,
		organizationID,
		checkoutID,
		Transition{
			Expected:        StatusProcessing,
			Status:          StatusProcessing,
			NextAction:      NextAction(session.NextAction),
			ProviderMessage: providerMessage,
		},
	)
	if err != nil {
		return TopupResult{}, err
	}

	return TopupResult{
		Checkout: checkoutRecord,
		Payment:  details.Payment,
		Session:  session,
	}, nil
}
