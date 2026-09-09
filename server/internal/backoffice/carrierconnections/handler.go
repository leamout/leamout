package carrierconnections

import (
	"bytes"
	"net/http"

	"github.com/a-h/templ"
)

type Handler struct {
	repository *Repository
}

func NewHandler(repository *Repository) *Handler {
	return &Handler{repository: repository}
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	connections, err := h.repository.List(r.Context())
	if err != nil {
		http.Error(w, "load carrier connections", http.StatusInternalServerError)
		return
	}
	render(w, r, Page(connections))
}

func render(w http.ResponseWriter, r *http.Request, view templ.Component) {
	var body bytes.Buffer
	if err := view.Render(r.Context(), &body); err != nil {
		http.Error(w, "render carrier connections view", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(body.Bytes())
}
