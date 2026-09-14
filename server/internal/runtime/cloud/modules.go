package server

import (
	"github.com/leamout/leamout/internal/commercial"
	"github.com/leamout/leamout/internal/identity"
	sharedmodules "github.com/leamout/leamout/internal/modules"
	"github.com/leamout/leamout/internal/platform/middleware"
	providerdiagnostics "github.com/leamout/leamout/internal/platform/provider_diagnostics"
	"github.com/leamout/leamout/internal/telecom"
	"github.com/leamout/leamout/internal/telecom/edge"
	"github.com/leamout/leamout/internal/telecom/wholesale"
	"github.com/leamout/leamout/internal/tenancy"
)

type Modules struct {
	Commercial           *commercial.Module
	Identity             *identity.Module
	Tenancy              *tenancy.Module
	Shared               *sharedmodules.Module
	Telecom              *telecom.Module
	Edge                 EdgeModule
	Wholesale            WholesaleModule
	ProviderDiagnostics  ProviderDiagnosticsModule
	RateLimit            *middleware.RateLimitMiddleware
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
