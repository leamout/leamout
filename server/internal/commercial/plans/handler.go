package plans

import (
	"net/http"

	"github.com/leamout/leamout/server/pkg/httputil"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	values, err := h.service.List(r.Context())
	if err != nil {
		httputil.Error(w, err)
		return
	}
	out := make([]Response, 0, len(values))
	for _, value := range values {
		out = append(out, response(value))
	}
	httputil.OK(w, map[string]any{"plans": out})
}
