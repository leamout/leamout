package identity

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/identity/auth"
	"github.com/leamout/leamout/server/internal/identity/session"
	"github.com/leamout/leamout/server/internal/identity/users"
	"github.com/leamout/leamout/server/internal/integrations/kinde"
	"github.com/leamout/leamout/server/internal/security/encryption"
)

type Module struct {
	Auth    AuthModule
	Session SessionModule
	Users   UsersModule
}

type AuthModule struct {
	Handler *auth.Handler
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

func New(
	queries *sqlc.Queries,
	development bool,
	db *pgxpool.Pool,
	client *kinde.Client,
	cipher *encryption.Cipher,
	frontendURL string,
) *Module {
	sessionRepository := session.NewRepository(queries)
	sessionService := session.NewService(sessionRepository, db, client, cipher)

	usersRepository := users.NewRepository(queries)
	usersService := users.NewService(usersRepository)

	return &Module{
		Auth: AuthModule{
			Handler: auth.NewHandler(queries, sessionService, client, cipher, frontendURL, development),
		},
		Session: SessionModule{
			Repository: sessionRepository,
			Service:    sessionService,
			Handler: session.NewHandler(
				sessionService,
				development,
				"",
			),
		},
		Users: UsersModule{
			Repository: usersRepository,
			Service:    usersService,
			Handler:    users.NewHandler(usersService),
		},
	}
}
