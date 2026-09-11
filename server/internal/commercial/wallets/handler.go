package wallets

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/helper"
	"github.com/leamout/leamout/pkg/httputil"
)

type TopupRequest struct {
	AmountMinor int64
	Provider    string
	Email       string
	CallbackURL string
	MobileMoney *MobileMoney
}

type MobileMoney struct {
	Phone    string `json:"phone"`
	Provider string `json:"provider"`
}

type TopupResponse struct {
	CheckoutID      uuid.UUID `json:"checkout_id"`
	PaymentID       uuid.UUID `json:"payment_id"`
	Reference       string    `json:"reference"`
	Provider        string    `json:"provider"`
	AmountMinor     int64     `json:"amount_minor"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	NextAction      string    `json:"next_action"`
	ProviderMessage *string   `json:"provider_message,omitempty"`
	ClientSecret    string    `json:"client_secret,omitempty"`
}

type TopupCreator func(context.Context, uuid.UUID, uuid.UUID, TopupRequest) (TopupResponse, error)

type Handler struct{ createTopup TopupCreator }

func NewHandler(createTopup TopupCreator) *Handler { return &Handler{createTopup: createTopup} }

type topupRequest struct {
	AmountMinor int64        `json:"amount_minor"`
	Provider    string       `json:"provider"`
	Email       string       `json:"email"`
	CallbackURL string       `json:"callback_url"`
	MobileMoney *MobileMoney `json:"mobile_money,omitempty"`
}

func (h *Handler) CreateTopup(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, apperror.NewBadRequest("organization context required"))
		return
	}
	walletID, err := uuid.Parse(chi.URLParam(r, "wallet_id"))
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid wallet_id"))
		return
	}
	request, err := helper.DecodeJSON[topupRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	result, err := h.createTopup(r.Context(), organizationID, walletID, TopupRequest{
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
	httputil.Created(w, result)
}
