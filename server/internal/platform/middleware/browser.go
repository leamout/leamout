package middleware

import (
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/leamout/leamout/server/pkg/httputil"
)

// Browser protects API writes from untrusted browser origins. Clients without
// browser origin headers can continue to authenticate through the API directly.
func Browser(frontendURL string) func(http.Handler) http.Handler {
	trustedOrigin := ""

	if parsed, err := url.Parse(frontendURL); err == nil && parsed.Host != "" {
		trustedOrigin = parsed.Scheme + "://" + parsed.Host
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-store")

			switch r.Method {
			case http.MethodGet, http.MethodHead, http.MethodOptions:
				next.ServeHTTP(w, r)

				return
			}

			origin := r.Header.Get("Origin")

			if (origin != "" && origin != trustedOrigin) ||
				(origin == "" && r.Header.Get("Sec-Fetch-Site") == "cross-site") {
				httputil.JSON(w, http.StatusForbidden, httputil.Response{
					Error: &httputil.ErrorObj{
						Code:    "FORBIDDEN",
						Message: "Request origin is not allowed",
					},
				})

				return
			}

			if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1/auth/") {
				mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))

				if err != nil || mediaType != "application/json" {
					httputil.JSON(w, http.StatusUnsupportedMediaType, httputil.Response{
						Error: &httputil.ErrorObj{
							Code:    "UNSUPPORTED_MEDIA_TYPE",
							Message: "Content-Type must be application/json",
						},
					})

					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
