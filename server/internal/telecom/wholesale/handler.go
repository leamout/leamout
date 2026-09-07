package wholesale

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"strings"

	"github.com/leamout/leamout/pkg/helper"
	"github.com/leamout/leamout/pkg/httputil"
)

type Handler struct {
	service *Service
	secret  string
}

func NewHandler(service *Service, secret string) *Handler {
	return &Handler{service: service, secret: strings.TrimSpace(secret)}
}

func (h *Handler) Reconcile(w http.ResponseWriter, r *http.Request) {
	provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if h.secret == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(h.secret)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	cdr, err := helper.DecodeJSON[CDR](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	result, err := h.service.Reconcile(r.Context(), cdr)
	if err != nil {
		switch {
		case errors.Is(err, ErrCallNotFound):
			http.Error(w, "managed call not found", http.StatusNotFound)
		case errors.Is(err, ErrCDRConflict):
			http.Error(w, "provider CDR conflict", http.StatusConflict)
		case errors.Is(err, ErrInvalidCDR):
			http.Error(w, "invalid provider CDR", http.StatusBadRequest)
		default:
			http.Error(w, "provider CDR reconciliation unavailable", http.StatusServiceUnavailable)
		}
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httputil.OK(w, result)
}
