package state

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/httputil"
)

// Handler exposes the resolved commercial state for an organization.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetOrganizationState(w http.ResponseWriter, r *http.Request) {
	organizationID, err := requestOrganizationID(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	resolved, err := h.service.Resolve(r.Context(), organizationID)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.OK(w, newOrganizationStateResponse(resolved))
}

func requestOrganizationID(r *http.Request) (uuid.UUID, error) {
	organizationID, err := uuid.Parse(chi.URLParam(r, "organization_id"))
	if err != nil {
		return uuid.Nil, apperror.NewBadRequest("invalid organization_id")
	}
	contextOrganizationID, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, apperror.NewBadRequest("organization context required")
	}
	if organizationID != contextOrganizationID {
		return uuid.Nil, apperror.NewForbidden("organization context does not match organization_id")
	}
	return organizationID, nil
}
