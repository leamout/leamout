package commercial

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
	id, err := uuid.Parse(chi.URLParam(r, "organization_id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	account, err := h.repository.Get(r.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "load commercial account", 500)
		return
	}
	render(w, r, DetailPage(account))
}

func NewHandler(repository *Repository) *Handler {
	return &Handler{repository: repository}
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.repository.List(r.Context())
	if err != nil {
		http.Error(w, "load commercial accounts", http.StatusInternalServerError)
		return
	}
	render(w, r, Page(accounts))
}

func render(w http.ResponseWriter, r *http.Request, view templ.Component) {
	var body bytes.Buffer
	if err := view.Render(r.Context(), &body); err != nil {
		http.Error(w, "render commercial view", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(body.Bytes())
}
