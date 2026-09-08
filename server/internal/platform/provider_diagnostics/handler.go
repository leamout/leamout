package provider_diagnostics

import (
	"crypto/subtle"
	"net/http"
	"strconv"
	"strings"

	"github.com/leamout/leamout/pkg/httputil"
)

type Handler struct {
	service *Service
	secret  string
}

func NewHandler(service *Service, secret string) *Handler {
	return &Handler{service: service, secret: strings.TrimSpace(secret)}
}

func (h *Handler) Snapshot(w http.ResponseWriter, r *http.Request) {
	provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if h.secret == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(h.secret)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	limit := int32(50)
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || value <= 0 {
			http.Error(w, "invalid limit", http.StatusBadRequest)
			return
		}
		limit = int32(value)
	}

	snapshot, err := h.service.Snapshot(r.Context(), limit)
	if err != nil {
		http.Error(w, "provider diagnostics unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httputil.OK(w, snapshot)
}
