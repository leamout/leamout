package identity

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/leamout/leamout/server/internal/identity/auth"
	"github.com/leamout/leamout/server/internal/identity/kindesync"
	"github.com/leamout/leamout/server/internal/identity/session"
	"github.com/leamout/leamout/server/internal/identity/users"
)

func RegisterRoutes(
	router chi.Router,
	module *Module,
	requireSession func(http.Handler) http.Handler,
) {
	kindesync.RegisterRoutes(router, module.KindeSync)

	auth.RegisterRoutes(
		router,
		module.Auth.Handler,
	)

	session.RegisterRoutes(
		router,
		module.Session.Handler,
		requireSession,
	)

	users.RegisterRoutes(
		router,
		module.Users.Handler,
		requireSession,
	)
}
