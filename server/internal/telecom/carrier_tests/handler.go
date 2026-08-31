package carrier_tests

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/helper"
	"github.com/leamout/leamout/pkg/httputil"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, apperror.NewBadRequest("organization context required"))
		return
	}
	connectionID, err := uuid.Parse(chi.URLParam(r, "carrier_connection_id"))
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid carrier_connection_id"))
		return
	}
	req, err := helper.DecodeJSON[Request](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	result, err := h.service.Run(r.Context(), organizationID, connectionID, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.Created(w, result)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, apperror.NewBadRequest("organization context required"))
		return
	}
	connectionID, err := uuid.Parse(chi.URLParam(r, "carrier_connection_id"))
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid carrier_connection_id"))
		return
	}
	limit, offset := int32(50), int32(0)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, parseErr := strconv.ParseInt(raw, 10, 32)
		if parseErr != nil {
			httputil.Error(w, apperror.NewBadRequest("invalid limit"))
			return
		}
		limit = int32(value)
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		value, parseErr := strconv.ParseInt(raw, 10, 32)
		if parseErr != nil {
			httputil.Error(w, apperror.NewBadRequest("invalid offset"))
			return
		}
		offset = int32(value)
	}
	items, err := h.service.List(r.Context(), organizationID, connectionID, limit, offset)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, map[string]any{"carrier_test_calls": items})
}
