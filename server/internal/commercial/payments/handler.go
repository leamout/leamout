package payments

import (
	"context"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/httputil"
)

const maxWebhookBytes = 1 << 20

type settlementCompleter interface {
	CompletePayment(context.Context, Settlement) error
}

type Handler struct {
	service    *Service
	completion settlementCompleter
}

func NewHandler(service *Service, completion settlementCompleter) *Handler {
	return &Handler{service: service, completion: completion}
}

func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid webhook payload"))
		return
	}

	event, err := h.service.ParseWebhook(providerName, payload, r.Header)
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid payment webhook"))
		return
	}

	settlement, err := h.service.ProcessProviderEvent(r.Context(), event)
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid payment webhook"))
		return
	}
	if h.completion != nil && isCompletableSettlement(settlement.Status) {
		if err = h.completion.CompletePayment(r.Context(), settlement); err != nil {
			httputil.Error(w, apperror.NewServiceUnavailable("complete checkout settlement", err))
			return
		}
	}
	httputil.OK(w, map[string]bool{"received": true})
}

func isCompletableSettlement(status Status) bool {
	return status == StatusSucceeded || status == StatusFailed || status == StatusCancelled
}
