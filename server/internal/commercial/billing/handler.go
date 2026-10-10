package billing

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/platform/middleware"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/pkg/httputil"
)

const maxWebhookBodyBytes = 1 << 20

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok || organizationID == uuid.Nil {
		httputil.Error(w, apperror.NewBadRequest("organization context required"))
		return
	}

	var req CheckoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid request body"))
		return
	}
	result, err := h.service.Checkout(r.Context(), organizationID, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, result)
}

func (h *Handler) Portal(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok || organizationID == uuid.Nil {
		httputil.Error(w, apperror.NewBadRequest("organization context required"))
		return
	}

	var req PortalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid request body"))
		return
	}
	result, err := h.service.Portal(r.Context(), organizationID, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, result)
}

func (h *Handler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBodyBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid Stripe webhook body"))
		return
	}
	if err := h.service.Webhook(r.Context(), payload, r.Header.Get("Stripe-Signature")); err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, map[string]bool{"received": true})
}
