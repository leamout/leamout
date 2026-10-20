package auth

import (
	"bytes"
	"net/http"

	identityauth "github.com/leamout/leamout/server/internal/identity/auth"
	"github.com/leamout/leamout/server/internal/identity/session"
	"github.com/leamout/leamout/server/internal/security/authn"
)

// Handler serves the standalone operator sign-in page and browser logout.
// Keycloak login and callback reuse the existing identity implementation.
type Handler struct {
	identity    *identityauth.Handler
	sessions    *session.Service
	development bool
}

func NewHandler(identity *identityauth.Handler, sessions *session.Service, development bool) *Handler {
	return &Handler{
		identity:    identity,
		sessions:    sessions,
		development: development,
	}
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

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(authn.SessionCookieName)
	if err == nil {
		stored, err := h.sessions.Get(r.Context(), cookie.Value)
		if err == nil {
			if err := h.sessions.Revoke(r.Context(), stored.ID, stored.UserID); err != nil {
				http.Error(w, "Unable to sign out", http.StatusServiceUnavailable)
				return
			}
		}
	}
	session.ClearCookie(w, h.development, "")
	http.Redirect(w, r, "/log-in", http.StatusSeeOther)
}
