package numbers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/httputil"
	"net/http"
)

type Handler struct{ service *Service }

func NewHandler(s *Service) *Handler { return &Handler{s} }
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	org, e := orgID(r)
	if e != nil {
		httputil.Error(w, e)
		return
	}
	var req CreateRequest
	if e := decode(r, &req); e != nil {
		httputil.Error(w, e)
		return
	}
	v, e := h.service.Create(r.Context(), org, req)
	if e != nil {
		httputil.Error(w, e)
		return
	}
	httputil.Created(w, response(v))
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	org, e := orgID(r)
	if e != nil {
		httputil.Error(w, e)
		return
	}
	v, e := h.service.List(r.Context(), org)
	if e != nil {
		httputil.Error(w, e)
		return
	}
	items := make([]Response, 0, len(v))
	for _, x := range v {
		items = append(items, response(x))
	}
	httputil.OK(w, map[string]any{"numbers": items})
}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	org, id, e := ids(r)
	if e != nil {
		httputil.Error(w, e)
		return
	}
	v, e := h.service.Get(r.Context(), org, id)
	if e != nil {
		httputil.Error(w, e)
		return
	}
	httputil.OK(w, response(v))
}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	org, id, e := ids(r)
	if e != nil {
		httputil.Error(w, e)
		return
	}
	var req UpdateRequest
	if e := decode(r, &req); e != nil {
		httputil.Error(w, e)
		return
	}
	v, e := h.service.Update(r.Context(), org, id, req)
	if e != nil {
		httputil.Error(w, e)
		return
	}
	httputil.OK(w, response(v))
}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	org, id, e := ids(r)
	if e != nil {
		httputil.Error(w, e)
		return
	}
	if e := h.service.Delete(r.Context(), org, id); e != nil {
		httputil.Error(w, e)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
func orgID(r *http.Request) (uuid.UUID, error) {
	v, ok := middleware.OrganizationIDFromContext(r.Context())
	if !ok {
		return uuid.Nil, apperror.NewBadRequest("organization context required")
	}
	return v, nil
}
func ids(r *http.Request) (uuid.UUID, uuid.UUID, error) {
	org, e := orgID(r)
	if e != nil {
		return uuid.Nil, uuid.Nil, e
	}
	id, e := uuid.Parse(chi.URLParam(r, "number_id"))
	if e != nil {
		return uuid.Nil, uuid.Nil, apperror.NewBadRequest("invalid number_id")
	}
	return org, id, nil
}
func decode(r *http.Request, target any) error {
	if e := json.NewDecoder(r.Body).Decode(target); e != nil {
		return apperror.NewBadRequest("invalid request body")
	}
	return nil
}
