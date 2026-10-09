package entitlements

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/platform/middleware"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/pkg/httputil"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok || organizationID == uuid.Nil {
		httputil.Error(w, apperror.NewBadRequest("organization context required"))
		return
	}
	values, err := h.service.ListEffective(r.Context(), organizationID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, map[string]any{"entitlements": values})
}
