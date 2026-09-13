package server

import (
	"github.com/leamout/leamout/internal/commercial"
	"github.com/leamout/leamout/internal/identity"
	"github.com/leamout/leamout/internal/modules/audit"
	"github.com/leamout/leamout/internal/modules/idempotency"
	"github.com/leamout/leamout/internal/modules/webhooks"
	"github.com/leamout/leamout/internal/platform/middleware"
	providerdiagnostics "github.com/leamout/leamout/internal/platform/provider_diagnostics"
	"github.com/leamout/leamout/internal/telecom/calls"
	"github.com/leamout/leamout/internal/telecom/carriers"
	"github.com/leamout/leamout/internal/telecom/conferences"
	"github.com/leamout/leamout/internal/telecom/edge"
	"github.com/leamout/leamout/internal/telecom/numbers"
	"github.com/leamout/leamout/internal/telecom/realtime"
	"github.com/leamout/leamout/internal/telecom/recordings"
	"github.com/leamout/leamout/internal/telecom/routing"
	"github.com/leamout/leamout/internal/telecom/sip_domains"
	"github.com/leamout/leamout/internal/telecom/subscribers"
	"github.com/leamout/leamout/internal/telecom/trunks"
	"github.com/leamout/leamout/internal/telecom/voice"
	"github.com/leamout/leamout/internal/telecom/wholesale"
	"github.com/leamout/leamout/internal/tenancy"
)

type Modules struct {
	Commercial           *commercial.Module
	Identity             *identity.Module
	Tenancy              *tenancy.Module
	Voice                VoiceModule
	Calls                CallsModule
	Recordings           RecordingsModule
	Conferences          ConferencesModule
	Webhooks             WebhooksModule
	Audit                AuditModule
	Idempotency          IdempotencyModule
	RateLimit            *middleware.RateLimitMiddleware
	SIPDomains           SIPDomainsModule
	Numbers              NumbersModule
	Subscribers          SubscribersModule
	Trunks               TrunksModule
	Carriers             CarriersModule
	Realtime             RealtimeModule
	Edge                 EdgeModule
	Routing              *routing.Service
	Wholesale            WholesaleModule
	ProviderDiagnostics  ProviderDiagnosticsModule
	Authn                *middleware.AuthnMiddleware
	OrganizationsContext *middleware.OrganizationMiddleware
}

type ProviderDiagnosticsModule struct {
	Repository *providerdiagnostics.Repository
	Service    *providerdiagnostics.Service
	Handler    *providerdiagnostics.Handler
}

type WholesaleModule struct {
	Repository *wholesale.Repository
	Service    *wholesale.Service
	Handler    *wholesale.Handler
}

type EdgeModule struct {
	Repository *edge.Repository
	Service    *edge.Service
	Handler    *edge.Handler
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

type IdempotencyModule struct {
	Repository *idempotency.Repository
	Service    *idempotency.Service
	Middleware *middleware.IdempotencyMiddleware
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
