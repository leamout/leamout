package auth

import (
	"bytes"
	"net/http"

	identityauth "github.com/leamout/leamout/server/internal/identity/auth"
)

// Handler serves the standalone operator sign-in page and Kinde login callbacks.
type Handler struct {
	identity *identityauth.Handler
}

func NewHandler(identity *identityauth.Handler) *Handler {
	return &Handler{identity: identity}
}

func (h *Handler) loginPage(w http.ResponseWriter, r *http.Request) {
	var body bytes.Buffer
	if err := LoginPage().Render(r.Context(), &body); err != nil {
		http.Error(w, "Unable to render sign-in page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(body.Bytes())
}
