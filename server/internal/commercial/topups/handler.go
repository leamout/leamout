package topups

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/checkout"
	paymentprovider "github.com/leamout/leamout/internal/integrations/payments"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/helper"
	"github.com/leamout/leamout/pkg/httputil"
)

const maxWebhookBytes = 1 << 20

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type createRequest struct {
	AmountMinor int64                        `json:"amount_minor"`
	Provider    checkout.Provider            `json:"provider"`
	Email       string                       `json:"email"`
	CallbackURL string                       `json:"callback_url"`
	MobileMoney *paymentprovider.MobileMoney `json:"mobile_money,omitempty"`
}

type continueRequest struct {
	Action paymentprovider.NextAction `json:"action"`
	Value  string                     `json:"value"`
}

type checkoutResponse struct {
	OrderID         uuid.UUID           `json:"order_id"`
	PaymentID       uuid.UUID           `json:"payment_id"`
	Reference       string              `json:"reference"`
	Provider        checkout.Provider   `json:"provider"`
	AmountMinor     int64               `json:"amount_minor"`
	Currency        string              `json:"currency"`
	Status          checkout.Status     `json:"status"`
	NextAction      checkout.NextAction `json:"next_action"`
	ProviderMessage *string             `json:"provider_message,omitempty"`
	ClientSecret    string              `json:"client_secret,omitempty"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	organizationID, walletID, err := requestIDs(r, "wallet_id")
	if err != nil {
		httputil.Error(w, err)
		return
	}
	request, err := helper.DecodeJSON[createRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	result, err := h.service.Create(r.Context(), organizationID, walletID, CreateInput{
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
	httputil.Created(w, responseFromCheckout(result))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	organizationID, orderID, err := requestIDs(r, "checkout_order_id")
	if err != nil {
		httputil.Error(w, err)
		return
	}
	result, err := h.service.Get(r.Context(), organizationID, orderID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, checkoutResponse{
		OrderID: result.Order.ID, PaymentID: result.Payment.ID,
		Reference: result.Order.Reference, Provider: result.Order.Provider,
		AmountMinor: result.Order.AmountMinor, Currency: result.Order.Currency,
		Status: result.Order.Status, NextAction: result.Order.NextAction,
		ProviderMessage: result.Order.ProviderMessage,
	})
}

func (h *Handler) Continue(w http.ResponseWriter, r *http.Request) {
	organizationID, orderID, err := requestIDs(r, "checkout_order_id")
	if err != nil {
		httputil.Error(w, err)
		return
	}
	request, err := helper.DecodeJSON[continueRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	result, err := h.service.Continue(r.Context(), organizationID, orderID, ContinueInput{
		Action: request.Action,
		Value:  request.Value,
	})
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, responseFromCheckout(result))
}

func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid webhook payload"))
		return
	}
	if _, err = h.service.Webhook(r.Context(), provider, payload, r.Header); err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid payment webhook"))
		return
	}
	httputil.OK(w, map[string]bool{"received": true})
}

func requestIDs(r *http.Request, resourceParam string) (uuid.UUID, uuid.UUID, error) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, uuid.Nil, apperror.NewBadRequest("organization context required")
	}
	resourceID, err := uuid.Parse(chi.URLParam(r, resourceParam))
	if err != nil {
		return uuid.Nil, uuid.Nil, apperror.NewBadRequest("invalid " + resourceParam)
	}
	return organizationID, resourceID, nil
}

func responseFromCheckout(result Checkout) checkoutResponse {
	return checkoutResponse{
		OrderID: result.Order.ID, PaymentID: result.Payment.ID,
		Reference: result.Order.Reference, Provider: result.Order.Provider,
		AmountMinor: result.Order.AmountMinor, Currency: result.Order.Currency,
		Status: result.Order.Status, NextAction: result.Order.NextAction,
		ProviderMessage: result.Order.ProviderMessage, ClientSecret: result.Session.ClientSecret,
	}
}
