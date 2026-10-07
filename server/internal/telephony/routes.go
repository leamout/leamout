package telephony

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/leamout/leamout/server/internal/telephony/calls"
	"github.com/leamout/leamout/server/internal/telephony/numbers"
	"github.com/leamout/leamout/server/internal/telephony/recordings"
	"github.com/leamout/leamout/server/internal/telephony/trunks"
	"github.com/leamout/leamout/server/internal/telephony/webrtc"
)

func RegisterRoutes(
	router chi.Router,
	module *Module,
	organizationAccess func(string) func(http.Handler) http.Handler,
	idempotency func(http.Handler) http.Handler,
) {
	calls.RegisterRoutes(
		router,
		module.Calls.Handler,
		organizationAccess("calls"),
		idempotency,
	)

	numbers.RegisterRoutes(
		router,
		module.Numbers.Handler,
		organizationAccess("numbers"),
		idempotency,
	)

	recordings.RegisterRoutes(
		router,
		module.Recordings.Handler,
		organizationAccess("recordings"),
	)

	trunks.RegisterRoutes(
		router,
		module.Trunks.Handler,
		organizationAccess("trunks"),
	)

	webrtc.RegisterRoutes(
		router,
		module.WebRTC.Handler,
		organizationAccess("webrtc"),
	)
}
