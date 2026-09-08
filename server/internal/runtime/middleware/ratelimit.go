package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ulule/limiter/v3"

	"github.com/leamout/leamout/pkg/apperror"
	"github.com/leamout/leamout/pkg/httputil"
)

// RateLimitMiddleware enforces a shared request budget for each organization.
// It is applied after authentication and organization resolution so callers
// cannot choose or spoof the identity used as the rate-limit key.
type RateLimitMiddleware struct {
	limiter *limiter.Limiter
}

func NewRateLimitMiddleware(store limiter.Store) (*RateLimitMiddleware, error) {
	if store == nil {
		return nil, fmt.Errorf("rate limit store is required")
	}

	return &RateLimitMiddleware{
		limiter: limiter.New(store, limiter.Rate{
			Period: time.Minute,
			Limit:  1000,
		}),
	}, nil
}

func (m *RateLimitMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		organizationID, ok := OrganizationIDFromContext(r.Context())
		if !ok {
			httputil.Error(w, apperror.NewUnauthorized("organization context required"))
			return
		}

		limit, err := m.limiter.Get(r.Context(), "http:organization:"+organizationID.String())
		if err != nil {
			httputil.Error(w, apperror.NewServiceUnavailable("rate limit service unavailable", err))
			return
		}

		w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(limit.Limit, 10))
		w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(limit.Remaining, 10))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(limit.Reset, 10))

		if limit.Reached {
			retryAfter := limit.Reset - time.Now().Unix()
			if retryAfter < 1 {
				retryAfter = 1
			}
			w.Header().Set("Retry-After", strconv.FormatInt(retryAfter, 10))
			httputil.Error(w, apperror.NewTooManyRequests("organization rate limit exceeded"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
