package wallets

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/checkout"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
)

type walletStore interface {
	Get(context.Context, uuid.UUID, uuid.UUID) (Wallet, error)
}

type checkoutStore interface {
	Create(context.Context, uuid.UUID, checkout.CreateInput) (checkout.Checkout, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (checkout.Checkout, error)
	ClaimRefresh(context.Context, uuid.UUID, uuid.UUID, time.Time) (checkout.Checkout, error)
	Transition(context.Context, uuid.UUID, uuid.UUID, checkout.Transition) (checkout.Checkout, error)
}

type paymentStore interface {
	Create(context.Context, uuid.UUID, string, commercialpayments.CreateInput) (commercialpayments.Payment, error)
	GetByCheckout(context.Context, uuid.UUID, uuid.UUID) (commercialpayments.Payment, error)
	SetProviderID(context.Context, uuid.UUID, uuid.UUID, string, commercialpayments.Status) (commercialpayments.Payment, error)
	UpdateStatus(context.Context, uuid.UUID, uuid.UUID, commercialpayments.Status, *time.Time) (commercialpayments.Payment, error)
}

type settlementStore interface {
	Reconcile(context.Context, paymentprovider.Event) (TopupSettlement, error)
}

type TopupService struct {
	wallets     walletStore
	checkouts   checkoutStore
	payments    paymentStore
	settlements settlementStore
	providers   map[string]paymentprovider.Provider
	now         func() time.Time
}

func NewTopupService(
	wallets walletStore,
	checkouts checkoutStore,
	payments paymentStore,
	settlements settlementStore,
	providers map[string]paymentprovider.Provider,
) *TopupService {
	return &TopupService{
		wallets:     wallets,
		checkouts:   checkouts,
		payments:    payments,
		settlements: settlements,
		providers:   providers,
		now:         time.Now,
	}
}

func (s *TopupService) SetProvider(name string, provider paymentprovider.Provider) {
	if provider != nil {
		s.providers[name] = provider
	}
}

func (s *TopupService) Create(
	ctx context.Context,
	organizationID uuid.UUID,
	walletID uuid.UUID,
	input TopupCreateInput,
) (TopupCheckout, error) {
	providerName := string(input.Provider)
	provider, ok := s.providers[providerName]
	if !ok {
		return TopupCheckout{}, ErrProviderUnavailable
	}

	wallet, err := s.wallets.Get(ctx, organizationID, walletID)
	if err != nil {
		return TopupCheckout{}, err
	}
	if wallet.Status != StatusActive || input.AmountMinor <= 0 || strings.TrimSpace(input.Email) == "" {
		return TopupCheckout{}, ErrInvalidTopup
	}

	method := checkout.MethodCard
	if input.Provider == checkout.ProviderPaystack {
		method = checkout.MethodMobileMoney
		if wallet.Currency != "GHS" || input.MobileMoney == nil {
			return TopupCheckout{}, ErrInvalidTopup
		}
	} else if input.Provider != checkout.ProviderStripe || input.MobileMoney != nil {
		return TopupCheckout{}, ErrInvalidTopup
	}

	now := s.now().UTC()
	reference := "topup." + uuid.NewString()
	checkoutRecord, err := s.checkouts.Create(ctx, organizationID, checkout.CreateInput{
		WalletID:      &walletID,
		Type:          checkout.TypeWalletTopup,
		Provider:      input.Provider,
		PaymentMethod: method,
		Reference:     reference,
		AmountMinor:   input.AmountMinor,
		Currency:      wallet.Currency,
		ExpiresAt:     now.Add(30 * time.Minute),
	})
	if err != nil {
		return TopupCheckout{}, err
	}

	payment, err := s.payments.Create(ctx, organizationID, providerName, commercialpayments.CreateInput{
		CheckoutID:  checkoutRecord.ID,
		Status:      commercialpayments.StatusPending,
		AmountMinor: checkoutRecord.AmountMinor,
		Currency:    checkoutRecord.Currency,
	})
	if err != nil {
		return TopupCheckout{}, err
	}

	session, err := provider.CreateCheckout(ctx, paymentprovider.CheckoutRequest{
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
		return TopupCheckout{}, err
	}
	if session.Provider != providerName || session.Reference != reference ||
		(input.Provider == checkout.ProviderStripe && session.ProviderID == "") {
		s.failPendingCheckout(ctx, organizationID, checkoutRecord, payment)
		return TopupCheckout{}, ErrPaymentMismatch
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
		return TopupCheckout{}, err
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
		checkout.Transition{
			Expected:        checkout.StatusPending,
			Status:          checkout.StatusProcessing,
			NextAction:      checkout.NextAction(session.NextAction),
			ProviderMessage: providerMessage,
		},
	)
	if err != nil {
		return TopupCheckout{}, err
	}

	return TopupCheckout{
		Checkout: checkoutRecord,
		Payment:  payment,
		Session:  session,
		Order:    checkoutRecord,
	}, nil
}

func (s *TopupService) failPendingCheckout(
	ctx context.Context,
	organizationID uuid.UUID,
	checkoutRecord checkout.Checkout,
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
		checkout.Transition{
			Expected:    checkout.StatusPending,
			Status:      checkout.StatusFailed,
			NextAction:  checkout.ActionNone,
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

	if checkoutRecord.Provider == checkout.ProviderPaystack && checkoutRecord.Status == checkout.StatusProcessing {
		claimed, claimErr := s.checkouts.ClaimRefresh(
			ctx,
			organizationID,
			checkoutID,
			s.now().UTC().Add(-10*time.Second),
		)
		if claimErr == nil {
			if refreshed, ok := s.refreshPaystack(ctx, claimed); ok {
				if _, err = s.settlements.Reconcile(ctx, refreshed); err != nil {
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
		} else if !errors.Is(claimErr, checkout.ErrCheckoutNotFound) {
			return TopupDetails{}, claimErr
		}
	}

	return TopupDetails{
		Checkout: checkoutRecord,
		Payment:  payment,
		Order:    checkoutRecord,
	}, nil
}

func (s *TopupService) refreshPaystack(
	ctx context.Context,
	checkoutRecord checkout.Checkout,
) (paymentprovider.Event, bool) {
	provider, ok := s.providers["paystack"]
	if !ok {
		return paymentprovider.Event{}, false
	}

	payment, err := provider.GetPayment(ctx, checkoutRecord.Reference)
	if err != nil || (payment.Status != paymentprovider.StatusSucceeded && payment.Status != paymentprovider.StatusFailed) {
		return paymentprovider.Event{}, false
	}

	raw, _ := json.Marshal(map[string]string{
		"source":    "charge_lookup",
		"reference": checkoutRecord.Reference,
		"status":    string(payment.Status),
	})

	eventType := "charge.failed"
	if payment.Status == paymentprovider.StatusSucceeded {
		eventType = "charge.success"
	}

	return paymentprovider.Event{
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
) (TopupCheckout, error) {
	details, err := s.Get(ctx, organizationID, checkoutID)
	if err != nil {
		return TopupCheckout{}, err
	}

	provider, ok := s.providers[string(details.Checkout.Provider)]
	if !ok {
		return TopupCheckout{}, ErrProviderUnavailable
	}

	continuation, ok := provider.(paymentprovider.ContinuationProvider)
	if !ok || details.Checkout.Provider != checkout.ProviderPaystack ||
		details.Checkout.Status != checkout.StatusProcessing {
		return TopupCheckout{}, ErrInvalidTopup
	}

	session, err := continuation.ContinueCheckout(ctx, paymentprovider.ContinueCheckoutRequest{
		Reference: details.Checkout.Reference,
		Action:    input.Action,
		Value:     input.Value,
	})
	if err != nil {
		return TopupCheckout{}, err
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
		checkout.Transition{
			Expected:        checkout.StatusProcessing,
			Status:          checkout.StatusProcessing,
			NextAction:      checkout.NextAction(session.NextAction),
			ProviderMessage: providerMessage,
		},
	)
	if err != nil {
		return TopupCheckout{}, err
	}

	return TopupCheckout{
		Checkout: checkoutRecord,
		Payment:  details.Payment,
		Session:  session,
		Order:    checkoutRecord,
	}, nil
}

func (s *TopupService) Webhook(
	ctx context.Context,
	providerName string,
	payload []byte,
	headers http.Header,
) (TopupSettlement, error) {
	provider, ok := s.providers[providerName]
	if !ok {
		return TopupSettlement{}, ErrProviderUnavailable
	}

	event, err := provider.ParseWebhook(payload, headers)
	if err != nil {
		return TopupSettlement{}, err
	}
	if event.Provider != providerName {
		return TopupSettlement{}, ErrPaymentMismatch
	}
	if !isTopupPaymentEvent(event.Provider, event.Type) {
		return TopupSettlement{}, nil
	}

	return s.settlements.Reconcile(ctx, event)
}

func isTopupPaymentEvent(provider, eventType string) bool {
	switch provider {
	case "paystack":
		return eventType == "charge.success" || eventType == "charge.failed"
	case "stripe":
		return eventType == "checkout.session.completed" || eventType == "checkout.session.expired"
	default:
		return false
	}
}
