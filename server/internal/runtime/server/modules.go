package server

import (
	"github.com/leamout/leamout/internal/identity/auth"
	"github.com/leamout/leamout/internal/identity/session"
	"github.com/leamout/leamout/internal/identity/users"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/conferences"
	"github.com/leamout/leamout/internal/telecom/recordings"
	"github.com/leamout/leamout/internal/telecom/voice"
	"github.com/leamout/leamout/internal/tenancy/credentials"
	"github.com/leamout/leamout/internal/tenancy/members"
	"github.com/leamout/leamout/internal/tenancy/organization"
)

type Modules struct {
	Auth                 AuthModule
	Session              SessionModule
	Users                UsersModule
	Organizations        OrganizationModule
	Members              MembersModule
	Credentials          CredentialsModule
	Voice                VoiceModule
	Calls                CallsModule
	Recordings           RecordingsModule
	Conferences          ConferencesModule
	Authn                *middleware.AuthnMiddleware
	OrganizationsContext *middleware.OrganizationMiddleware
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

type CredentialsModule struct {
	Repository *credentials.Repository
	Service    *credentials.Service
	Handler    *credentials.Handler
}

type VoiceModule struct {
	Repository *voice.Repository
	Service    *voice.Service
	Handler    *voice.Handler
}

type CallsModule struct {
	Repository *calls.Repository
	Service    *calls.Service
	Handler    *calls.Handler
}

type RecordingsModule struct {
	Repository *recordings.Repository
	Service    *recordings.Service
	Handler    *recordings.Handler
}

type ConferencesModule struct {
	Repository *conferences.Repository
	Service    *conferences.Service
	Handler    *conferences.Handler
}
