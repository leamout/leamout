package numbers

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Handler struct {
	repository *Repository
}

func (h *Handler) detail(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "number_id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	number, err := h.repository.Get(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "load phone number", http.StatusInternalServerError)
		return
	}
	render(w, r, DetailPage(number))
}

func NewHandler(repository *Repository) *Handler {
	return &Handler{repository: repository}
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	numbers, err := h.repository.List(r.Context())
	if err != nil {
		http.Error(w, "load phone numbers", http.StatusInternalServerError)
		return
	}
	render(w, r, Page(numbers))
}

func render(w http.ResponseWriter, r *http.Request, view templ.Component) {
	var body bytes.Buffer
	if err := view.Render(r.Context(), &body); err != nil {
		http.Error(w, "render numbers view", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(body.Bytes())
}
