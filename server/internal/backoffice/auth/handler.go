package auth

import (
	"bytes"
	"errors"
	"net/http"
	"time"

	"github.com/a-h/templ"

	"github.com/leamout/leamout/internal/security/authn"
	"github.com/leamout/leamout/pkg/helper"
)

const CookieName = "leamout-backoffice-session"

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) loginPage(w http.ResponseWriter, r *http.Request) {
	renderLogin(w, r, http.StatusOK, LoginPageData{})
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	form, err := parseLoginForm(r)
	if err != nil {
		renderLogin(w, r, http.StatusBadRequest, LoginPageData{
			Email: form.Email,
			Error: "Email and password are required.",
		})
		return
	}

	token, sess, err := h.service.LoginWithPassword(
		r.Context(),
		form.Email,
		form.Password,
		helper.ClientIP(r),
		helper.UserAgent(r),
	)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		renderLogin(w, r, http.StatusUnauthorized, LoginPageData{
			Email: form.Email,
			Error: "Invalid email or password.",
		})
		return
	}

	SetCookie(w, token, sess.ExpiresAt.Time)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	principal, ok := authn.PrincipalFromContext(r.Context())
	if ok {
		_ = h.service.Logout(r.Context(), principal)
	}

	ClearCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func SetCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

func renderLogin(w http.ResponseWriter, r *http.Request, status int, data LoginPageData) {
	var body bytes.Buffer
	if err := render(LoginPage(data), r, &body); err != nil {
		http.Error(w, "render Backoffice login", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(body.Bytes())
}

func render(view templ.Component, r *http.Request, body *bytes.Buffer) error {
	return view.Render(r.Context(), body)
}
