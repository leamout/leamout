package limits

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/platform/middleware"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/pkg/httputil"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok || id == uuid.Nil {
		httputil.Error(w, apperror.NewBadRequest("organization context required"))
		return
	}
	value, err := h.service.Get(r.Context(), id)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, value)
}
func (h *Handler) RuntimeUsage(w http.ResponseWriter, r *http.Request) {
	id, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok || id == uuid.Nil {
		httputil.Error(w, apperror.NewBadRequest("organization context required"))
		return
	}
	value, err := h.service.RuntimeUsage(r.Context(), id)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, value)
}
