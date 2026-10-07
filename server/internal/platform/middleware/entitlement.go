package middleware

import (
	"net/http"

	"github.com/leamout/leamout/server/internal/platform/entitlements"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/pkg/httputil"
)

type EntitlementMiddleware struct {
	service *entitlements.Service
}

func NewEntitlementMiddleware(service *entitlements.Service) *EntitlementMiddleware {
	return &EntitlementMiddleware{
		service: service,
	}
}

func (m *EntitlementMiddleware) Require(capability entitlements.Capability) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			organizationID, ok := OrganizationIDFromContext(r.Context())
			if !ok {
				httputil.Error(w, apperror.NewBadRequest("organization context required"))
				return
			}
			if err := m.service.Require(r.Context(), organizationID, capability); err != nil {
				httputil.Error(w, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
