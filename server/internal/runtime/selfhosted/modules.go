package server

import (
	"github.com/leamout/leamout/internal/identity"
	sharedmodules "github.com/leamout/leamout/internal/modules"
	"github.com/leamout/leamout/internal/platform/middleware"
	"github.com/leamout/leamout/internal/telecom"
	"github.com/leamout/leamout/internal/tenancy"
)

type Modules struct {
	Identity             *identity.Module
	Tenancy              *tenancy.Module
	Shared               *sharedmodules.Module
	Telecom              *telecom.Module
	RateLimit            *middleware.RateLimitMiddleware
	Authn                *middleware.AuthnMiddleware
	OrganizationsContext *middleware.OrganizationMiddleware
}
