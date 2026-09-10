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
	Create(context.Context, uuid.UUID, checkout.CreateInput) (checkout.Order, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (checkout.Order, error)
	ClaimRefresh(context.Context, uuid.UUID, uuid.UUID, time.Time) (checkout.Order, error)
	Transition(context.Context, uuid.UUID, uuid.UUID, checkout.Transition) (checkout.Order, error)
}

type paymentStore interface {
	Create(context.Context, uuid.UUID, string, commercialpayments.CreateInput) (commercialpayments.Payment, error)
	GetByCheckoutOrder(context.Context, uuid.UUID, uuid.UUID) (commercialpayments.Payment, error)
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
	order, err := s.checkouts.Create(ctx, organizationID, checkout.CreateInput{
		WalletID:      &walletID,
		Type:          checkout.OrderWalletTopup,
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
		CheckoutOrderID: order.ID,
		Status:          commercialpayments.StatusPending,
		AmountMinor:     order.AmountMinor,
		Currency:        order.Currency,
	})
	if err != nil {
		return TopupCheckout{}, err
	}

	session, err := provider.CreateCheckout(ctx, paymentprovider.CheckoutRequest{
		Reference:   reference,
		AmountMinor: order.AmountMinor,
		Currency:    order.Currency,
		Email:       strings.TrimSpace(input.Email),
		CallbackURL: strings.TrimSpace(input.CallbackURL),
		Metadata: map[string]string{
			"organization_id": organizationID.String(),
			"wallet_id":       walletID.String(),
		},
		MobileMoney: input.MobileMoney,
	})
	if err != nil {
		s.failPendingCheckout(ctx, organizationID, order, payment)
		return TopupCheckout{}, err
	}
	if session.Provider != providerName || session.Reference != reference ||
		(input.Provider == checkout.ProviderStripe && session.ProviderID == "") {
		s.failPendingCheckout(ctx, organizationID, order, payment)
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

	order, err = s.checkouts.Transition(ctx, organizationID, order.ID, checkout.Transition{
		Expected:        checkout.StatusPending,
		Status:          checkout.StatusProcessing,
		NextAction:      checkout.NextAction(session.NextAction),
		ProviderMessage: providerMessage,
	})
	if err != nil {
		return TopupCheckout{}, err
	}

	return TopupCheckout{
		Order:   order,
		Payment: payment,
		Session: session,
	}, nil
}

func (s *TopupService) failPendingCheckout(
	ctx context.Context,
	organizationID uuid.UUID,
	order checkout.Order,
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
	_, _ = s.checkouts.Transition(ctx, organizationID, order.ID, checkout.Transition{
		Expected:    checkout.StatusPending,
		Status:      checkout.StatusFailed,
		NextAction:  checkout.ActionNone,
		CompletedAt: &completedAt,
	})
}

func (s *TopupService) Get(
	ctx context.Context,
	organizationID uuid.UUID,
	orderID uuid.UUID,
) (TopupDetails, error) {
	order, err := s.checkouts.Get(ctx, organizationID, orderID)
	if err != nil {
		return TopupDetails{}, err
	}

	payment, err := s.payments.GetByCheckoutOrder(ctx, organizationID, orderID)
	if err != nil {
		return TopupDetails{}, err
	}

	if order.Provider == checkout.ProviderPaystack && order.Status == checkout.StatusProcessing {
		claimed, claimErr := s.checkouts.ClaimRefresh(
			ctx,
			organizationID,
			orderID,
			s.now().UTC().Add(-10*time.Second),
		)
		if claimErr == nil {
			if refreshed, ok := s.refreshPaystack(ctx, claimed); ok {
				if _, err = s.settlements.Reconcile(ctx, refreshed); err != nil {
					return TopupDetails{}, err
				}

				order, err = s.checkouts.Get(ctx, organizationID, orderID)
				if err != nil {
					return TopupDetails{}, err
				}

				payment, err = s.payments.GetByCheckoutOrder(ctx, organizationID, orderID)
				if err != nil {
					return TopupDetails{}, err
				}
			}
		} else if !errors.Is(claimErr, checkout.ErrOrderNotFound) {
			return TopupDetails{}, claimErr
		}
	}

	return TopupDetails{
		Order:   order,
		Payment: payment,
	}, nil
}

func (s *TopupService) refreshPaystack(
	ctx context.Context,
	order checkout.Order,
) (paymentprovider.Event, bool) {
	provider, ok := s.providers["paystack"]
	if !ok {
		return paymentprovider.Event{}, false
	}

	payment, err := provider.GetPayment(ctx, order.Reference)
	if err != nil || (payment.Status != paymentprovider.StatusSucceeded && payment.Status != paymentprovider.StatusFailed) {
		return paymentprovider.Event{}, false
	}

	raw, _ := json.Marshal(map[string]string{
		"source":    "charge_lookup",
		"reference": order.Reference,
		"status":    string(payment.Status),
	})

	eventType := "charge.failed"
	if payment.Status == paymentprovider.StatusSucceeded {
		eventType = "charge.success"
	}

	return paymentprovider.Event{
		Provider:        "paystack",
		ProviderEventID: "lookup:" + order.Reference + ":" + string(payment.Status),
		Type:            eventType,
		Payment:         payment,
		Raw:             raw,
	}, true
}

func (s *TopupService) Continue(
	ctx context.Context,
	organizationID uuid.UUID,
	orderID uuid.UUID,
	input TopupContinueInput,
) (TopupCheckout, error) {
	details, err := s.Get(ctx, organizationID, orderID)
	if err != nil {
		return TopupCheckout{}, err
	}

	provider, ok := s.providers[string(details.Order.Provider)]
	if !ok {
		return TopupCheckout{}, ErrProviderUnavailable
	}

	continuation, ok := provider.(paymentprovider.ContinuationProvider)
	if !ok || details.Order.Provider != checkout.ProviderPaystack || details.Order.Status != checkout.StatusProcessing {
		return TopupCheckout{}, ErrInvalidTopup
	}

	session, err := continuation.ContinueCheckout(ctx, paymentprovider.ContinueCheckoutRequest{
		Reference: details.Order.Reference,
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

	order, err := s.checkouts.Transition(ctx, organizationID, orderID, checkout.Transition{
		Expected:        checkout.StatusProcessing,
		Status:          checkout.StatusProcessing,
		NextAction:      checkout.NextAction(session.NextAction),
		ProviderMessage: providerMessage,
	})
	if err != nil {
		return TopupCheckout{}, err
	}

	return TopupCheckout{
		Order:   order,
		Payment: details.Payment,
		Session: session,
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
