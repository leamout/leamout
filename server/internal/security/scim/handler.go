package scim

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	platformmiddleware "github.com/leamout/leamout/server/internal/platform/middleware"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/pkg/helper"
	"github.com/leamout/leamout/server/pkg/httputil"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}
func (h *Handler) CreateToken(w http.ResponseWriter, r *http.Request) {
	organizationID, err := organizationID(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	req, err := helper.DecodeJSON[CreateTokenRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	value, err := h.service.CreateToken(r.Context(), organizationID, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.Created(w, value)
}
func (h *Handler) ListTokens(w http.ResponseWriter, r *http.Request) {
	organizationID, err := organizationID(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	values, err := h.service.ListTokens(r.Context(), organizationID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, map[string]any{
		"scim_tokens": values,
	})
}
func (h *Handler) RevokeToken(w http.ResponseWriter, r *http.Request) {
	organizationID, err := organizationID(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "scim_token_id"))
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid scim_token_id"))
		return
	}
	if _, err := h.service.RevokeToken(r.Context(), organizationID, id); err != nil {
		httputil.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func organizationID(r *http.Request) (uuid.UUID, error) {
	id, ok := platformmiddleware.OrganizationIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, apperror.NewBadRequest("organization context required")
	}
	return id, nil
}
