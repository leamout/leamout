package server

import (
	"github.com/leamout/leamout/internal/identity/auth"
	"github.com/leamout/leamout/internal/identity/session"
	"github.com/leamout/leamout/internal/identity/users"
	"github.com/leamout/leamout/internal/modules/audit"
	"github.com/leamout/leamout/internal/modules/webhooks"
	"github.com/leamout/leamout/internal/runtime/middleware"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/carriers"
	"github.com/leamout/leamout/internal/telecom/conferences"
	"github.com/leamout/leamout/internal/telecom/numbers"
	"github.com/leamout/leamout/internal/telecom/realtime"
	"github.com/leamout/leamout/internal/telecom/recordings"
	"github.com/leamout/leamout/internal/telecom/sip_domains"
	"github.com/leamout/leamout/internal/telecom/subscribers"
	"github.com/leamout/leamout/internal/telecom/trunks"
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
	Webhooks             WebhooksModule
	Audit                AuditModule
	SIPDomains           SIPDomainsModule
	Numbers              NumbersModule
	Subscribers          SubscribersModule
	Trunks               TrunksModule
	Carriers             CarriersModule
	Realtime             RealtimeModule
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

type WebhooksModule struct {
	Repository *webhooks.Repository
	Service    *webhooks.Service
	Handler    *webhooks.Handler
}

type AuditModule struct {
	Repository *audit.Repository
	Service    *audit.Service
	Handler    *audit.Handler
}

type SIPDomainsModule struct {
	Repository *sip_domains.Repository
	Service    *sip_domains.Service
	Handler    *sip_domains.Handler
}

type NumbersModule struct {
	Repository *numbers.Repository
	Service    *numbers.Service
	Handler    *numbers.Handler
}

type SubscribersModule struct {
	Repository *subscribers.Repository
	Service    *subscribers.Service
	Handler    *subscribers.Handler
}

type TrunksModule struct {
	Repository *trunks.Repository
	Service    *trunks.Service
	Handler    *trunks.Handler
}

type CarriersModule struct {
	Repository *carriers.Repository
	Service    *carriers.Service
	Handler    *carriers.Handler
}

type RealtimeModule struct {
	Service *realtime.Service
	Handler *realtime.Handler
}
