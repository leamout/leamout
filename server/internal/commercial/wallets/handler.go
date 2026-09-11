package wallets

import (
	"context"
	"net/http"
	"time"

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

type walletReader interface {
	List(context.Context, uuid.UUID) ([]Wallet, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (Wallet, error)
	Balance(context.Context, uuid.UUID, uuid.UUID) (Balance, error)
	ListEntries(context.Context, uuid.UUID, uuid.UUID) ([]LedgerEntry, error)
}

type Handler struct {
	wallets     walletReader
	createTopup TopupCreator
}

func NewHandler(wallets walletReader, createTopup TopupCreator) *Handler {
	return &Handler{wallets: wallets, createTopup: createTopup}
}

type walletResponse struct {
	ID             uuid.UUID        `json:"id"`
	OrganizationID uuid.UUID        `json:"organization_id"`
	Currency       string           `json:"currency"`
	Status         Status           `json:"status"`
	Balance        *balanceResponse `json:"balance,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

type balanceResponse struct {
	PostedMinor    int64 `json:"posted_minor"`
	ReservedMinor  int64 `json:"reserved_minor"`
	AvailableMinor int64 `json:"available_minor"`
}

type ledgerEntryResponse struct {
	ID             uuid.UUID `json:"id"`
	WalletID       uuid.UUID `json:"wallet_id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Type           EntryType `json:"entry_type"`
	AmountMinor    int64     `json:"amount_minor"`
	SourceType     string    `json:"source_type"`
	SourceID       string    `json:"source_id"`
	OccurredAt     time.Time `json:"occurred_at"`
	CreatedAt      time.Time `json:"created_at"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	organizationID, err := requestContextOrganizationID(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	items, err := h.wallets.List(r.Context(), organizationID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	responses := make([]walletResponse, 0, len(items))
	for _, wallet := range items {
		balance, balanceErr := h.wallets.Balance(r.Context(), organizationID, wallet.ID)
		if balanceErr != nil {
			httputil.Error(w, balanceErr)
			return
		}
		responses = append(responses, newWalletResponse(wallet, &balance))
	}
	httputil.OK(w, map[string]any{"wallets": responses})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	organizationID, walletID, err := requestWallet(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	wallet, err := h.wallets.Get(r.Context(), organizationID, walletID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	balance, err := h.wallets.Balance(r.Context(), organizationID, walletID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, newWalletResponse(wallet, &balance))
}

func (h *Handler) ListLedger(w http.ResponseWriter, r *http.Request) {
	organizationID, walletID, err := requestWallet(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	if _, err = h.wallets.Get(r.Context(), organizationID, walletID); err != nil {
		httputil.Error(w, err)
		return
	}
	entries, err := h.wallets.ListEntries(r.Context(), organizationID, walletID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	responses := make([]ledgerEntryResponse, 0, len(entries))
	for _, entry := range entries {
		responses = append(responses, ledgerEntryResponse{
			ID: entry.ID, WalletID: entry.WalletID, OrganizationID: entry.OrganizationID,
			Type: entry.Type, AmountMinor: entry.AmountMinor, SourceType: entry.SourceType,
			SourceID: entry.SourceID, OccurredAt: entry.OccurredAt, CreatedAt: entry.CreatedAt,
		})
	}
	httputil.OK(w, map[string]any{"entries": responses})
}

func newWalletResponse(wallet Wallet, balance *Balance) walletResponse {
	response := walletResponse{
		ID: wallet.ID, OrganizationID: wallet.OrganizationID, Currency: wallet.Currency,
		Status: wallet.Status, CreatedAt: wallet.CreatedAt, UpdatedAt: wallet.UpdatedAt,
	}
	if balance != nil {
		response.Balance = &balanceResponse{
			PostedMinor: balance.PostedMinor, ReservedMinor: balance.ReservedMinor, AvailableMinor: balance.AvailableMinor,
		}
	}
	return response
}

type topupRequest struct {
	AmountMinor int64        `json:"amount_minor"`
	Provider    string       `json:"provider"`
	Email       string       `json:"email"`
	CallbackURL string       `json:"callback_url"`
	MobileMoney *MobileMoney `json:"mobile_money,omitempty"`
}

func (h *Handler) CreateTopup(w http.ResponseWriter, r *http.Request) {
	organizationID, walletID, err := requestWallet(r)
	if err != nil {
		httputil.Error(w, err)
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

func requestWallet(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	organizationID, err := requestContextOrganizationID(r)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	walletID, err := uuid.Parse(chi.URLParam(r, "wallet_id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, apperror.NewBadRequest("invalid wallet_id")
	}
	return organizationID, walletID, nil
}

func requestContextOrganizationID(r *http.Request) (uuid.UUID, error) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, apperror.NewBadRequest("organization context required")
	}
	return organizationID, nil
}
