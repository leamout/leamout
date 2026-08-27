package server

import (
	"github.com/leamout/leamout/internal/identity/auth"
	"github.com/leamout/leamout/internal/identity/session"
	"github.com/leamout/leamout/internal/identity/users"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/internal/tenancy/members"
	"github.com/leamout/leamout/internal/tenancy/organization"
)

type Modules struct {
	Auth          AuthModule
	Session       SessionModule
	Users         UsersModule
	Organizations OrganizationModule
	Members       MembersModule
	Authn         *middleware.AuthnMiddleware
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

type OrganizationModule struct {
	Repository *organization.Repository
	Service    *organization.Service
	Handler    *organization.Handler
}

type MembersModule struct {
	Repository *members.Repository
	Service    *members.Service
	Handler    *members.Handler
}
