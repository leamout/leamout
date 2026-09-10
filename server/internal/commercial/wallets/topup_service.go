package wallets

import (
	"context"
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

func NewTopupService(wallets walletStore, checkouts checkoutStore, payments paymentStore, settlements settlementStore, providers map[string]paymentprovider.Provider) *TopupService {
	return &TopupService{wallets: wallets, checkouts: checkouts, payments: payments, settlements: settlements, providers: providers, now: time.Now}
}

func (s *TopupService) SetProvider(name string, provider paymentprovider.Provider) {
	if provider != nil {
		s.providers[name] = provider
	}
}

func (s *TopupService) Create(ctx context.Context, organizationID, walletID uuid.UUID, input TopupCreateInput) (TopupCheckout, error) {
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
		WalletID: &walletID, Type: checkout.OrderWalletTopup, Provider: input.Provider,
		PaymentMethod: method, Reference: reference, AmountMinor: input.AmountMinor,
		Currency: wallet.Currency, ExpiresAt: now.Add(30 * time.Minute),
	})
	if err != nil {
		return TopupCheckout{}, err
	}
	payment, err := s.payments.Create(ctx, organizationID, providerName, commercialpayments.CreateInput{
		CheckoutOrderID: order.ID, Status: commercialpayments.StatusPending,
		AmountMinor: order.AmountMinor, Currency: order.Currency,
	})
	if err != nil {
		return TopupCheckout{}, err
	}

	session, err := provider.CreateCheckout(ctx, paymentprovider.CheckoutRequest{
		Reference: reference, AmountMinor: order.AmountMinor, Currency: order.Currency,
		Email: strings.TrimSpace(input.Email), CallbackURL: strings.TrimSpace(input.CallbackURL),
		Metadata:    map[string]string{"organization_id": organizationID.String(), "wallet_id": walletID.String()},
		MobileMoney: input.MobileMoney,
	})
	if err != nil {
		s.failPendingCheckout(ctx, organizationID, order, payment)
		return TopupCheckout{}, err
	}
	if session.Provider != providerName || session.ProviderID == "" || session.Reference != reference {
		s.failPendingCheckout(ctx, organizationID, order, payment)
		return TopupCheckout{}, ErrPaymentMismatch
	}
	payment, err = s.payments.SetProviderID(ctx, organizationID, payment.ID, session.ProviderID, commercialpayments.StatusProcessing)
	if err != nil {
		return TopupCheckout{}, err
	}
	message := strings.TrimSpace(session.Message)
	var providerMessage *string
	if message != "" {
		providerMessage = &message
	}
	order, err = s.checkouts.Transition(ctx, organizationID, order.ID, checkout.Transition{
		Expected: checkout.StatusPending, Status: checkout.StatusProcessing,
		NextAction: checkout.NextAction(session.NextAction), ProviderMessage: providerMessage,
	})
	if err != nil {
		return TopupCheckout{}, err
	}
	return TopupCheckout{Order: order, Payment: payment, Session: session}, nil
}

func (s *TopupService) failPendingCheckout(
	ctx context.Context,
	organizationID uuid.UUID,
	order checkout.Order,
	payment commercialpayments.Payment,
) {
	completedAt := s.now().UTC()
	_, _ = s.payments.UpdateStatus(ctx, organizationID, payment.ID, commercialpayments.StatusFailed, nil)
	_, _ = s.checkouts.Transition(ctx, organizationID, order.ID, checkout.Transition{
		Expected:    checkout.StatusPending,
		Status:      checkout.StatusFailed,
		NextAction:  checkout.ActionNone,
		CompletedAt: &completedAt,
	})
}

func (s *TopupService) Get(ctx context.Context, organizationID, orderID uuid.UUID) (TopupDetails, error) {
	order, err := s.checkouts.Get(ctx, organizationID, orderID)
	if err != nil {
		return TopupDetails{}, err
	}
	payment, err := s.payments.GetByCheckoutOrder(ctx, organizationID, orderID)
	if err != nil {
		return TopupDetails{}, err
	}
	return TopupDetails{Order: order, Payment: payment}, nil
}

func (s *TopupService) Continue(ctx context.Context, organizationID, orderID uuid.UUID, input TopupContinueInput) (TopupCheckout, error) {
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
		Reference: details.Order.Reference, Action: input.Action, Value: input.Value,
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
		Expected: checkout.StatusProcessing, Status: checkout.StatusProcessing,
		NextAction: checkout.NextAction(session.NextAction), ProviderMessage: providerMessage,
	})
	if err != nil {
		return TopupCheckout{}, err
	}
	return TopupCheckout{Order: order, Payment: details.Payment, Session: session}, nil
}

func (s *TopupService) Webhook(ctx context.Context, providerName string, payload []byte, headers http.Header) (TopupSettlement, error) {
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
	return s.settlements.Reconcile(ctx, event)
}
