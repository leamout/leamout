package audit

import (
	"net/http"
	"strconv"

	"github.com/leamout/leamout/server/internal/platform/middleware"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/pkg/httputil"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, apperror.NewBadRequest("organization context required"))
		return
	}
	req, filterErr := parseFilters(r)
	if filterErr != nil {
		httputil.Error(w, filterErr)
		return
	}
	limit, offset := int32(50), int32(0)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			httputil.Error(w, apperror.NewBadRequest("invalid limit"))
			return
		}
		limit = int32(value)
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			httputil.Error(w, apperror.NewBadRequest("invalid offset"))
			return
		}
		offset = int32(value)
	}
	req.Offset, req.Limit = offset, limit
	items, err := h.service.List(r.Context(), organizationID, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, map[string]any{"audit_events": items})
}
