package numbers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/helper"
	"github.com/leamout/leamout/pkg/httputil"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	organizationID, err := organizationID(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	req, err := helper.DecodeJSON[BYOCCreateRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	number, err := h.service.CreateBYOC(r.Context(), organizationID, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.Created(w, response(number))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	organizationID, err := organizationID(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	numbers, err := h.service.List(r.Context(), organizationID)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	items := make([]Response, 0, len(numbers))
	for _, number := range numbers {
		items = append(items, response(number))
	}

	httputil.OK(w, map[string]any{"numbers": items})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	organizationID, id, err := ids(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	number, err := h.service.Get(r.Context(), organizationID, id)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.OK(w, response(number))
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	organizationID, id, err := ids(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	req, err := helper.DecodeJSON[UpdateRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	number, err := h.service.Update(r.Context(), organizationID, id, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.OK(w, response(number))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	organizationID, id, err := ids(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	if err := h.service.ReleaseBYOC(r.Context(), organizationID, id); err != nil {
		httputil.Error(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SetCarrierConnection(w http.ResponseWriter, r *http.Request) {
	organizationID, id, err := ids(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	req, err := helper.DecodeJSON[CarrierConnectionRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	number, err := h.service.SetCarrierConnection(r.Context(), organizationID, id, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, response(number))
}

func organizationID(r *http.Request) (uuid.UUID, error) {
	id, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, apperror.NewBadRequest("organization context required")
	}

	return id, nil
}

func ids(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	organizationID, err := organizationID(r)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}

	id, err := uuid.Parse(chi.URLParam(r, "number_id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, apperror.NewBadRequest("invalid number_id")
	}

	return organizationID, id, nil
}
