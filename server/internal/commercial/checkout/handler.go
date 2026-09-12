package checkout

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/commercial/payments"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/helper"
	"github.com/leamout/leamout/pkg/httputil"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

type CreateRequest struct {
	WalletID    *uuid.UUID      `json:"wallet_id"`
	AmountMinor int64           `json:"amount_minor"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

type ConfirmRequest struct {
	PaymentMethod PaymentMethod         `json:"payment_method"`
	Email         string                `json:"email"`
	CallbackURL   string                `json:"callback_url"`
	MobileMoney   *payments.MobileMoney `json:"mobile_money,omitempty"`
}

type ContinueRequest struct {
	Action payments.NextAction `json:"action"`
	Value  string              `json:"value"`
}

type CheckoutResponse struct {
	CheckoutID      uuid.UUID       `json:"checkout_id"`
	PaymentID       *uuid.UUID      `json:"payment_id,omitempty"`
	Type            Type            `json:"type"`
	WalletID        *uuid.UUID      `json:"wallet_id,omitempty"`
	Reference       string          `json:"reference"`
	Provider        Provider        `json:"provider,omitempty"`
	PaymentMethod   PaymentMethod   `json:"payment_method,omitempty"`
	AmountMinor     int64           `json:"amount_minor"`
	Currency        string          `json:"currency"`
	Status          Status          `json:"status"`
	NextAction      NextAction      `json:"next_action"`
	ProviderMessage *string         `json:"provider_message,omitempty"`
	ClientSecret    string          `json:"client_secret,omitempty"`
	ExpiresAt       time.Time       `json:"expires_at"`
	CompletedAt     *time.Time      `json:"completed_at,omitempty"`
	Metadata        json.RawMessage `json:"metadata,omitempty"`
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
	checkoutRecord, err := h.service.Create(r.Context(), organizationID, CreateParams(request))
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.Created(w, Response(Result{Checkout: checkoutRecord}))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	organizationID, checkoutID, err := requestIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	result, err := h.service.Get(r.Context(), organizationID, checkoutID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, Response(result))
}

func (h *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	organizationID, checkoutID, err := requestIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	request, err := helper.DecodeJSON[ConfirmRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	result, err := h.service.Confirm(r.Context(), organizationID, checkoutID, ConfirmInput(request))
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, Response(result))
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
	result, err := h.service.Continue(r.Context(), organizationID, checkoutID, ContinueInput(request))
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

func Response(result Result) CheckoutResponse {
	response := CheckoutResponse{
		CheckoutID: result.Checkout.ID, Type: result.Checkout.Type, WalletID: result.Checkout.WalletID,
		Reference: result.Checkout.Reference, Provider: result.Checkout.Provider, PaymentMethod: result.Checkout.PaymentMethod,
		AmountMinor: result.Checkout.AmountMinor, Currency: result.Checkout.Currency, Status: result.Checkout.Status,
		NextAction: result.Checkout.NextAction, ProviderMessage: result.Checkout.ProviderMessage,
		ExpiresAt: result.Checkout.ExpiresAt, CompletedAt: result.Checkout.CompletedAt, Metadata: result.Checkout.Metadata,
	}
	if result.Payment != nil {
		id := result.Payment.ID
		response.PaymentID = &id
	}
	if result.Session != nil {
		response.ClientSecret = result.Session.ClientSecret
	}
	return response
}
