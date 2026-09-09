package calls

import (
	"bytes"
	"net/http"

	"github.com/a-h/templ"
)

type Handler struct {
	repository *Repository
	calls      []Call
}

func NewHandler(repository *Repository) *Handler {
	return &Handler{
		repository: repository,
		calls: []Call{
			{ID: "call_01JQ8YN7", From: "+1 415 555 0100", To: "+1 212 555 0198", Direction: "Outbound", Duration: "03:42", Status: "Completed"},
			{ID: "call_01JQ8XKV", From: "+44 20 7946 0958", To: "+1 415 555 0100", Direction: "Inbound", Duration: "00:18", Status: "In progress"},
		},
	}
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	render(w, r, Page(h.calls))
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
