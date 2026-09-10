package calls

import (
	"bytes"
	"context"
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type repository interface {
	List(context.Context) ([]Call, error)
	Get(context.Context, uuid.UUID) (Detail, error)
}

type Handler struct {
	repository repository
}

func NewHandler(repository repository) *Handler {
	return &Handler{repository: repository}
}

func (h *Handler) detail(w http.ResponseWriter, r *http.Request) {
	callID, err := uuid.Parse(chi.URLParam(r, "call_id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	call, err := h.repository.Get(r.Context(), callID)
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "load call", http.StatusInternalServerError)
		return
	}
	render(w, r, DetailPage(call))
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	calls, err := h.repository.List(r.Context())
	if err != nil {
		http.Error(w, "load calls", http.StatusInternalServerError)
		return
	}
	render(w, r, Page(calls))
}

func render(w http.ResponseWriter, r *http.Request, view templ.Component) {
	var body bytes.Buffer
	if err := view.Render(r.Context(), &body); err != nil {
		http.Error(w, "render calls view", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(body.Bytes())
}
