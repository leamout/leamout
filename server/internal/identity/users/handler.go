package users

import (
	"net/http"

	"github.com/leamout/leamout/internal/security/authn"
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

func (h *Handler) Current(w http.ResponseWriter, r *http.Request) {
	userID, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, apperror.NewUnauthorized("authentication required"))
		return
	}

	user, err := h.service.Get(r.Context(), userID)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.OK(w, toResponse(user))
}

func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := authn.UserIDFromContext(r.Context())
	if !ok {
		httputil.Error(w, apperror.NewUnauthorized("authentication required"))
		return
	}

	req, err := helper.DecodeJSON[UpdateProfileRequest](r)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	if err := validateUpdateProfileRequest(req); err != nil {
		httputil.Error(w, err)
		return
	}

	user, err := h.service.UpdateProfile(r.Context(), userID, req.Name)
	if err != nil {
		httputil.Error(w, err)
		return
	}

	httputil.OK(w, toResponse(user))
}
