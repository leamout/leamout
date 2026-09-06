package edge

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

func (h *Handler) Admit(w http.ResponseWriter, r *http.Request) {
	provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if h.secret == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(h.secret)) != 1 {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	req, err := helper.DecodeJSON[Request](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	if req.Username == "" || req.Realm == "" || !e164(req.From) || !e164(req.To) {
		http.Error(w, "invalid admission request", http.StatusBadRequest)
		return
	}
	decision, err := h.service.Admit(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrDenied) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "admission unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httputil.JSON(w, http.StatusOK, decision)
}

func e164(value string) bool {
	if len(value) < 8 || len(value) > 16 || value[0] != '+' || value[1] < '1' || value[1] > '9' {
		return false
	}
	for i := 2; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}
