package users

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/leamout/leamout/server/internal/observability/logging"
)

type Handler struct {
	repository *Repository
	logger     *logging.Logger
}

func NewHandler(repository *Repository, logger *logging.Logger) *Handler {
	return &Handler{repository: repository, logger: logger}
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	filter, err := parseFilter(r.URL.Query())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	result, err := h.repository.List(r.Context(), filter)
	if err != nil {
		h.logger.Error(r.Context(), "list backoffice users", "error", err)
		http.Error(w, "Unable to load users", http.StatusInternalServerError)
		return
	}
	h.render(w, r, Page(result))
}

func (h *Handler) detail(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "user_id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	memberPage, err := parsePage(r.URL.Query().Get("members_page"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.repository.Get(r.Context(), id, memberPage)
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		h.logger.Error(r.Context(), "get backoffice user", "error", err)
		http.Error(w, "Unable to load user", http.StatusInternalServerError)
		return
	}
	h.render(w, r, DetailPage(user))
}

func (h *Handler) render(w http.ResponseWriter, r *http.Request, view templ.Component) {
	var body bytes.Buffer
	if err := view.Render(r.Context(), &body); err != nil {
		h.logger.Error(r.Context(), "render backoffice user", "error", err)
		http.Error(w, "Unable to render user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(body.Bytes())
}
