package invitations

import (
	"github.com/coffeyvidzro/monogo/internal/security/authn"
	"github.com/coffeyvidzro/monogo/pkg/apperror"
	"github.com/coffeyvidzro/monogo/pkg/helper"
	"github.com/coffeyvidzro/monogo/pkg/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	actor, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, apperror.NewUnauthorized("authentication required"))
		return
	}
	organizationID, err := uuid.Parse(chi.URLParam(r, "organization_id"))
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid organization_id"))
		return
	}
	req, err := helper.DecodeJSON[CreateRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	response, err := h.service.Create(r.Context(), actor, organizationID, req)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.Created(w, response)
}
func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	actor, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, apperror.NewUnauthorized("authentication required"))
		return
	}
	req, err := helper.DecodeJSON[AcceptRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	response, err := h.service.Accept(r.Context(), actor, req.Token)
	if err != nil {
		httputil.Error(w, err)
		return
	}
	httputil.OK(w, response)
}
func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	actor, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, apperror.NewUnauthorized("authentication required"))
		return
	}
	organizationID, err := uuid.Parse(chi.URLParam(r, "organization_id"))
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid organization_id"))
		return
	}
	invitationID, err := uuid.Parse(chi.URLParam(r, "invitation_id"))
	if err != nil {
		httputil.Error(w, apperror.NewBadRequest("invalid invitation_id"))
		return
	}
	if err := h.service.Revoke(r.Context(), actor, organizationID, invitationID); err != nil {
		httputil.Error(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
