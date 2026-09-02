package subscriptions

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/helper"
	"github.com/leamout/leamout/pkg/httputil"
)

// Handler exposes organization-scoped customer subscription operations.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	organizationID, err := requestOrganizationID(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	subscriptions, err := h.service.List(r.Context(), organizationID)
	if err != nil {
		writeSubscriptionError(w, err)
		return
	}

	responses := make([]subscriptionResponse, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		responses = append(responses, newSubscriptionResponse(subscription))
	}
	httputil.OK(w, map[string]any{"subscriptions": responses})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	organizationID, subscriptionID, err := requestSubscriptionIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	subscription, err := h.service.Get(r.Context(), organizationID, subscriptionID)
	if err != nil {
		writeSubscriptionError(w, err)
		return
	}

	httputil.OK(w, newSubscriptionResponse(subscription))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	organizationID, subscriptionID, err := requestSubscriptionIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	request, err := helper.DecodeJSON[UpdateRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	subscription, err := h.service.ChangePrice(r.Context(), organizationID, subscriptionID, request.PriceID)
	if err != nil {
		writeSubscriptionError(w, err)
		return
	}

	httputil.OK(w, newSubscriptionResponse(subscription))
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	organizationID, subscriptionID, err := requestSubscriptionIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	subscription, err := h.service.Cancel(r.Context(), organizationID, subscriptionID)
	if err != nil {
		writeSubscriptionError(w, err)
		return
	}

	httputil.OK(w, newSubscriptionResponse(subscription))
}

func requestOrganizationID(r *http.Request) (uuid.UUID, error) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, apperror.NewBadRequest("organization context required")
	}
	return organizationID, nil
}

func requestSubscriptionIDs(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	organizationID, err := requestOrganizationID(r)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	subscriptionID, err := uuid.Parse(chi.URLParam(r, "subscription_id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, apperror.NewBadRequest("invalid subscription_id")
	}
	return organizationID, subscriptionID, nil
}

func writeSubscriptionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrSubscriptionNotFound):
		httputil.Error(w, apperror.NewNotFound("subscription not found"))
	case errors.Is(err, ErrPriceUnavailable):
		httputil.Error(w, apperror.NewBadRequest("subscription price is unavailable"))
	case errors.Is(err, ErrPriceIDRequired):
		httputil.Error(w, apperror.NewBadRequest("price_id is required"))
	case errors.Is(err, ErrTerminalSubscription):
		httputil.Error(w, apperror.NewConflict("terminal subscription cannot change commercial terms"))
	case errors.Is(err, ErrInvalidTransition):
		httputil.Error(w, apperror.NewConflict("subscription cannot be cancelled from its current status"))
	case errors.Is(err, ErrOrganizationUnavailable):
		httputil.Error(w, apperror.NewConflict("organization is unavailable for subscription changes"))
	default:
		httputil.Error(w, err)
	}
}
