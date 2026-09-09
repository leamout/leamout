package auth

import "github.com/google/uuid"

type LoginForm struct {
	Email    string
	Password string
}

type LoginPageData struct {
	Email string
	Error string
}

type PlatformUser struct {
	ID              uuid.UUID
	IsPlatformAdmin bool
}
