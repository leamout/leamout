package carriers

import (
	"net/http"

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
	org, err := organizationID(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	req, err := helper.DecodeJSON[CreateRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	item, err := h.service.Create(r.Context(), org, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.Created(w, item)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	org, err := organizationID(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	items, err := h.service.List(r.Context(), org)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, map[string]any{"carrier_connections": items})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	org, id, err := connectionIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	item, err := h.service.Get(r.Context(), org, id)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, item)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	org, id, err := connectionIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	req, err := helper.DecodeJSON[UpdateRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	item, err := h.service.Update(r.Context(), org, id, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, item)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	org, id, err := connectionIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	if err := h.service.Delete(r.Context(), org, id); err != nil {
		httputil.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateSourceIP(w http.ResponseWriter, r *http.Request) {
	org, id, err := connectionIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	req, err := helper.DecodeJSON[SourceIPCreateRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	item, err := h.service.CreateSourceIP(r.Context(), org, id, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.Created(w, item)
}

func (h *Handler) ListSourceIPs(w http.ResponseWriter, r *http.Request) {
	org, id, err := connectionIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	items, err := h.service.ListSourceIPs(r.Context(), org, id)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, map[string]any{"source_ips": items})
}

func (h *Handler) DeleteSourceIP(w http.ResponseWriter, r *http.Request) {
	org, connectionID, err := connectionIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "source_ip_id"))
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid source_ip_id"))
		return
	}
	if err := h.service.DeleteSourceIP(r.Context(), org, connectionID, id); err != nil {
		httputil.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func organizationID(r *http.Request) (uuid.UUID, error) {
	id, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, apperror.NewBadRequest("organization context required")
	}
	return id, nil
}

func connectionIDs(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	org, err := organizationID(r)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	id, err := uuid.Parse(chi.URLParam(r, "carrier_connection_id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, apperror.NewBadRequest("invalid carrier_connection_id")
	}
	return org, id, nil
}
