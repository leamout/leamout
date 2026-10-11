package tenancy

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leamout/leamout/server/internal/database/sqlc"
	"github.com/leamout/leamout/server/internal/platform/email"
	"github.com/leamout/leamout/server/internal/tenancy/credentials"
	"github.com/leamout/leamout/server/internal/tenancy/invitations"
	"github.com/leamout/leamout/server/internal/tenancy/members"
	"github.com/leamout/leamout/server/internal/tenancy/organization"
)

type Module struct {
	Invitations   *invitations.Handler
	Organizations OrganizationModule
	Members       MembersModule
	Credentials   CredentialsModule
}

type OrganizationModule struct {
	Repository *organization.Repository
	Service    *organization.Service
	Handler    *organization.Handler
}

type MembersModule struct {
	Repository *members.Repository
	Service    *members.Service
	Handler    *members.Handler
}

type CredentialsModule struct {
	Repository *credentials.Repository
	Service    *credentials.Service
	Handler    *credentials.Handler
}

func New(queries *sqlc.Queries) *Module {
	organizations := organization.NewRepository(queries)
	organizationService := organization.NewService(organizations)

	membersRepository := members.NewRepository(queries)
	membersService := members.NewService(membersRepository)

	credentialsRepository := credentials.NewRepository(queries)
	credentialsService := credentials.NewService(credentialsRepository)

	return &Module{
		Organizations: OrganizationModule{
			Repository: organizations,
			Service:    organizationService,
			Handler:    organization.NewHandler(organizationService),
		},
		Members: MembersModule{
			Repository: membersRepository,
			Service:    membersService,
			Handler:    members.NewHandler(membersService),
		},
		Credentials: CredentialsModule{
			Repository: credentialsRepository,
			Service:    credentialsService,
			Handler:    credentials.NewHandler(credentialsService),
		},
	}
}

func (m *Module) ConfigureEmail(pool *pgxpool.Pool, emails *email.Service, acceptURL string) {
	repo := invitations.NewRepository(pool)
	service := invitations.NewService(repo, emails, acceptURL)
	m.Invitations = invitations.NewHandler(service)
}
