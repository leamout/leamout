package payments

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/httputil"
)

const maxWebhookBytes = 1 << 20

type Handler struct {
	service   *Service
	providers *ProviderRegistry
}

func NewHandler(service *Service, providers *ProviderRegistry) *Handler {
	return &Handler{service: service, providers: providers}
}

func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	providerName := chi.URLParam(r, "provider")
	provider, ok := h.providers.Get(providerName)
	if !ok {
		httputil.Error(w, apperror.NewBadRequest("invalid payment webhook"))
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxWebhookBytes)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid webhook payload"))
		return
	}
	event, err := provider.ParseWebhook(payload, r.Header)
	if err != nil || event.Provider != providerName {
		httputil.Error(w, apperror.NewBadRequest("invalid payment webhook"))
		return
	}
	if _, err = h.service.ProcessProviderEvent(r.Context(), event); err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid payment webhook"))
		return
	}
	httputil.OK(w, map[string]bool{"received": true})
}
