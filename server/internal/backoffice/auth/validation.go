package auth

import (
	"errors"
	"net/http"
	"strings"
)

var ErrInvalidLoginForm = errors.New("invalid login form")

func parseLoginForm(r *http.Request) (LoginForm, error) {
	if err := r.ParseForm(); err != nil {
		return LoginForm{}, ErrInvalidLoginForm
	}

	form := LoginForm{
		Email:    strings.TrimSpace(r.FormValue("email")),
		Password: r.FormValue("password"),
	}
	if form.Email == "" || form.Password == "" {
		return form, ErrInvalidLoginForm
	}
	return form, nil
}
