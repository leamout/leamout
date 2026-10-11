package invitations

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/server/internal/security/authn"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/pkg/helper"
	"github.com/leamout/leamout/server/pkg/httputil"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func organizationIDs(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	userID, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, uuid.Nil, apperror.NewUnauthorized("authentication required")
	}
	organizationID, err := uuid.Parse(chi.URLParam(r, "organization_id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, apperror.NewBadRequest("invalid organization_id")
	}
	return userID, organizationID, nil
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, organizationID, err := organizationIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	req, err := helper.DecodeJSON[CreateRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	result, err := h.service.Create(r.Context(), userID, organizationID, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.Created(w, result)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, organizationID, err := organizationIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	result, err := h.service.List(r.Context(), userID, organizationID)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, result)
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	userID, organizationID, err := organizationIDs(r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "invitation_id"))
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid invitation_id"))
		return
	}
	if err := h.service.Revoke(r.Context(), userID, organizationID, id); err != nil {
		httputil.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	userID, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, apperror.NewUnauthorized("authentication required"))
		return
	}
	req, err := helper.DecodeJSON[AcceptRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	if err := h.service.Accept(r.Context(), userID, req.Token); err != nil {
		httputil.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
