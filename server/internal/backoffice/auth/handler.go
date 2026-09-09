package auth

import (
	"errors"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/leamout/leamout/internal/security/authn"
	"github.com/leamout/leamout/pkg/helper"
)

const CookieName = "leamout-backoffice-session"

var loginTemplate = template.Must(template.New("backoffice-login").Parse(`<!doctype html>
<html lang="en" data-theme="corporate">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Sign in · Leamout Backoffice</title>
  <link rel="icon" href="/favicon.ico">
  <link rel="stylesheet" href="/static/css/tailwindcss.css">
</head>
<body class="min-h-screen bg-base-200 text-base-content">
  <main class="flex min-h-screen items-center justify-center p-6">
    <section class="card w-full max-w-md border border-base-300 bg-base-100 shadow-xl">
      <div class="card-body gap-5">
        <div>
          <p class="text-sm font-medium uppercase tracking-wide opacity-60">Leamout</p>
          <h1 class="card-title text-2xl">Backoffice</h1>
          <p class="mt-2 text-sm opacity-70">Sign in with your platform administrator account.</p>
        </div>
        {{if .Error}}
        <div class="alert alert-error" role="alert"><span>{{.Error}}</span></div>
        {{end}}
        <form method="post" action="/login" class="space-y-4">
          <label class="form-control w-full">
            <span class="label-text mb-1">Email</span>
            <input class="input input-bordered w-full" type="email" name="email" autocomplete="username" required autofocus>
          </label>
          <label class="form-control w-full">
            <span class="label-text mb-1">Password</span>
            <input class="input input-bordered w-full" type="password" name="password" autocomplete="current-password" required>
          </label>
          <button class="btn btn-primary w-full" type="submit">Sign in</button>
        </form>
      </div>
    </section>
  </main>
</body>
</html>`))

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) LoginPage(w http.ResponseWriter, _ *http.Request) {
	renderLogin(w, http.StatusOK, "")
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		renderLogin(w, http.StatusBadRequest, "Invalid login request.")
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")
	if email == "" || password == "" {
		renderLogin(w, http.StatusBadRequest, "Email and password are required.")
		return
	}

	token, sess, err := h.service.LoginWithPassword(
		r.Context(),
		email,
		password,
		helper.ClientIP(r),
		helper.UserAgent(r),
	)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		renderLogin(w, http.StatusUnauthorized, "Invalid email or password.")
		return
	}

	SetCookie(w, token, sess.ExpiresAt.Time)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
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

func renderLogin(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = loginTemplate.Execute(w, struct{ Error string }{Error: message})
}
