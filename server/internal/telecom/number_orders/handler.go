package number_orders

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

	req, err := helper.DecodeJSON[CreateRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	order, err := h.service.Create(r.Context(), organizationID, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.Created(w, order)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	organizationID, err := organizationID(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "number_order_id"))
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid number_order_id"))
		return
	}

	order, err := h.service.Get(r.Context(), organizationID, id)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.OK(w, order)
}

func organizationID(r *http.Request) (uuid.UUID, error) {
	id, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, apperror.NewBadRequest("organization context required")
	}
	return id, nil
}
