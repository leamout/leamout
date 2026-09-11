package topups

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	checkout "github.com/leamout/leamout/internal/commercial/checkout"
	commercialpayments "github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/helper"
	"github.com/leamout/leamout/pkg/httputil"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type createRequest struct {
	AmountMinor int64                           `json:"amount_minor"`
	Provider    checkout.Provider               `json:"provider"`
	Email       string                          `json:"email"`
	CallbackURL string                          `json:"callback_url"`
	MobileMoney *commercialpayments.MobileMoney `json:"mobile_money,omitempty"`
}

type continueRequest struct {
	Action commercialpayments.NextAction `json:"action"`
	Value  string                        `json:"value"`
}

type checkoutResponse struct {
	CheckoutID      uuid.UUID           `json:"checkout_id"`
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

	result, err := h.service.Create(
		r.Context(),
		organizationID,
		walletID,
		CreateInput(request),
	)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.Created(w, responseFromCheckout(result))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	organizationID, checkoutID, err := requestIDs(r, "checkout_id")
	if err != nil {
		httputil.Error(w, err)
		return
	}

	result, err := h.service.Get(r.Context(), organizationID, checkoutID)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.OK(w, checkoutResponse{
		CheckoutID:      result.Checkout.ID,
		PaymentID:       result.Payment.ID,
		Reference:       result.Checkout.Reference,
		Provider:        result.Checkout.Provider,
		AmountMinor:     result.Checkout.AmountMinor,
		Currency:        result.Checkout.Currency,
		Status:          result.Checkout.Status,
		NextAction:      result.Checkout.NextAction,
		ProviderMessage: result.Checkout.ProviderMessage,
	})
}

func (h *Handler) Continue(w http.ResponseWriter, r *http.Request) {
	organizationID, checkoutID, err := requestIDs(r, "checkout_id")
	if err != nil {
		httputil.Error(w, err)
		return
	}

	request, err := helper.DecodeJSON[continueRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	result, err := h.service.Continue(
		r.Context(),
		organizationID,
		checkoutID,
		ContinueInput(request),
	)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.OK(w, responseFromCheckout(result))
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
		CheckoutID:      result.Checkout.ID,
		PaymentID:       result.Payment.ID,
		Reference:       result.Checkout.Reference,
		Provider:        result.Checkout.Provider,
		AmountMinor:     result.Checkout.AmountMinor,
		Currency:        result.Checkout.Currency,
		Status:          result.Checkout.Status,
		NextAction:      result.Checkout.NextAction,
		ProviderMessage: result.Checkout.ProviderMessage,
		ClientSecret:    result.Session.ClientSecret,
	}
}
