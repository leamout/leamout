package server

import (
	"github.com/leamout/leamout/internal/identity/auth"
	"github.com/leamout/leamout/internal/identity/session"
	"github.com/leamout/leamout/internal/identity/users"
	"github.com/leamout/leamout/internal/runtime/middleware"
)

type Modules struct {
	Auth    AuthModule
	Session SessionModule
	Users   UsersModule
	Authn   *middleware.AuthnMiddleware
}

type AuthModule struct {
	Repository *auth.Repository
	Service    *auth.Service
	Handler    *auth.Handler
}

type SessionModule struct {
	Repository *session.Repository
	Service    *session.Service
	Handler    *session.Handler
}

type UsersModule struct {
	Repository *users.Repository
	Service    *users.Service
	Handler    *users.Handler
}
