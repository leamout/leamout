package entitlements

import (
	"net/http"

	"github.com/leamout/leamout/server/internal/platform/middleware"
	"github.com/leamout/leamout/server/pkg/apperror"
	"github.com/leamout/leamout/server/pkg/httputil"
)

type Middleware struct {
	service *Service
}

func NewMiddleware(service *Service) *Middleware {
	return &Middleware{service: service}
}

func (m *Middleware) Require(capability Capability) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			organizationID, ok := middleware.OrganizationIDFromContext(r.Context())
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
