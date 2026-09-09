package organizations

import (
	"bytes"
	"net/http"

	"github.com/a-h/templ"
)

type Handler struct {
	organizations []Organization
}

func NewHandler() *Handler {
	return &Handler{organizations: []Organization{
		{Name: "Acme Communications", Slug: "acme", Plan: "Scale", Members: 12, Status: "Active"},
		{Name: "Northstar Voice", Slug: "northstar", Plan: "Growth", Members: 5, Status: "Active"},
	}}
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	page := pageFrom(r.URL.Query().Get("page"))
	start := min((page-1)*pageSize, len(h.organizations))
	end := min(start+pageSize, len(h.organizations))
	render(w, r, Page(h.organizations[start:end], page, max(1, (len(h.organizations)+pageSize-1)/pageSize)))
}

func render(w http.ResponseWriter, r *http.Request, view templ.Component) {
	var body bytes.Buffer
	if err := view.Render(r.Context(), &body); err != nil {
		http.Error(w, "render organizations view", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(body.Bytes())
}
