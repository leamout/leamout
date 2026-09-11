package checkout

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/helper"
	"github.com/leamout/leamout/pkg/httputil"
)

type Handler struct{ topups *TopupService }

func NewHandler(topups *TopupService) *Handler { return &Handler{topups: topups} }

type ContinueRequest struct {
	Action payments.NextAction `json:"action"`
	Value  string              `json:"value"`
}

type CreateRequest struct {
	Type        Type                  `json:"type"`
	WalletID    uuid.UUID             `json:"wallet_id"`
	AmountMinor int64                 `json:"amount_minor"`
	Provider    Provider              `json:"provider"`
	Email       string                `json:"email"`
	CallbackURL string                `json:"callback_url"`
	MobileMoney *payments.MobileMoney `json:"mobile_money,omitempty"`
}

type CheckoutResponse struct {
	CheckoutID      uuid.UUID  `json:"checkout_id"`
	PaymentID       uuid.UUID  `json:"payment_id"`
	Reference       string     `json:"reference"`
	Provider        Provider   `json:"provider"`
	AmountMinor     int64      `json:"amount_minor"`
	Currency        string     `json:"currency"`
	Status          Status     `json:"status"`
	NextAction      NextAction `json:"next_action"`
	ProviderMessage *string    `json:"provider_message,omitempty"`
	ClientSecret    string     `json:"client_secret,omitempty"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, apperror.NewBadRequest("organization context required"))
		return
	}
	request, err := helper.DecodeJSON[CreateRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	if request.Type != TypeWalletTopup || request.WalletID == uuid.Nil {
		httputil.Error(w, ErrInvalidCheckout)
		return
	}
	result, err := h.topups.Create(r.Context(), organizationID, request.WalletID, TopupCreateInput{
		AmountMinor: request.AmountMinor,
		Provider:    request.Provider,
		Email:       request.Email,
		CallbackURL: request.CallbackURL,
		MobileMoney: request.MobileMoney,
	})
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.Created(w, Response(result))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	organizationID, checkoutID, err := requestIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	result, err := h.topups.Get(r.Context(), organizationID, checkoutID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, Response(TopupResult{Checkout: result.Checkout, Payment: result.Payment}))
}

func (h *Handler) Continue(w http.ResponseWriter, r *http.Request) {
	organizationID, checkoutID, err := requestIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	request, err := helper.DecodeJSON[ContinueRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	result, err := h.topups.Continue(r.Context(), organizationID, checkoutID, TopupContinueInput(request))
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, Response(result))
}

func requestIDs(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, uuid.Nil, apperror.NewBadRequest("organization context required")
	}
	checkoutID, err := uuid.Parse(chi.URLParam(r, "checkout_id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, apperror.NewBadRequest("invalid checkout_id")
	}
	return organizationID, checkoutID, nil
}

func Response(result TopupResult) CheckoutResponse {
	return CheckoutResponse{CheckoutID: result.Checkout.ID, PaymentID: result.Payment.ID, Reference: result.Checkout.Reference, Provider: result.Checkout.Provider, AmountMinor: result.Checkout.AmountMinor, Currency: result.Checkout.Currency, Status: result.Checkout.Status, NextAction: result.Checkout.NextAction, ProviderMessage: result.Checkout.ProviderMessage, ClientSecret: result.Session.ClientSecret}
}
