# Leamout Backoffice

Backoffice is a standalone full-stack application built with Go, Templ, HTMX,
TailwindCSS, DaisyUI and Hyperscript. The Go process serves HTML and embedded
assets directly on port 8082. It does not require the Next.js console.

Organizations provide organization name/ID search, lifecycle filters,
25-row pagination and detail pages with resource counts and subscription state.
Disabled and deleted organizations remain inspectable.

Agents provide platform-wide name/ID/organization-name search, status and engine
filters, an exact organization filter, and 25-row pagination. Detail pages show
instructions, voice settings, policies, configuration revision and the owning
organization. Disabled agents and agents belonging to disabled or deleted
organizations remain inspectable. These pages are read-only and exclude raw
engine configuration and provider credentials.

## Development

Backoffice uses `config.Load()` and `config.Config` from
`internal/platform/config/config.go`, loading the shared `server/.env`. Supply
all required shared application settings. Backoffice connects to PostgreSQL and
Keycloak; it does not start the API, workers or telephony services.

Set `BACKOFFICE_URL` to the public origin of this application (default
`http://localhost:8082`). For local HTTP, use `APP_ENV=development` and a local
Keycloak issuer. The configured Keycloak client must allow the exact redirect
URI `${BACKOFFICE_URL}/auth/callback`. The realm template includes this URI
alongside the existing API application's callback; existing realms require
updating the allowed redirect URIs explicitly.

From the repository root, with Go from `server/go.mod` and Bun 1.3.14 installed:

```sh
make backoffice-assets backoffice-generate
cd server
go run ./cmd/backoffice
```

Open `http://localhost:8082/log-in`. The Templ page starts the Keycloak flow at
`/auth/login`; `/auth/callback` creates a host-only browser session and returns
the operator to `/organizations`. The flow reuses the identity package's state,
nonce, PKCE and encrypted session implementation. It requires no console login,
console cookie or console proxy.

The sign-out form posts to `/auth/logout`, revokes the current local session,
clears its cookie and redirects to `/log-in`. Keycloak's SSO session remains
available for subsequent sign-in. Signed backchannel logout events can be
received at `/identity/backchannel-logout`; the existing API backchannel handler
also revokes matching sessions in the shared database.

Only active users with `is_platform_admin` may access organization and agent pages.
A trusted administrator must explicitly promote an existing, verified account
using its exact ID:

```sql
UPDATE users
SET is_platform_admin = TRUE, updated_at = NOW()
WHERE id = '<verified-operator-user-id>' AND disabled_at IS NULL;
```

Removing the flag, disabling the user or revoking the session denies subsequent
protected requests. No default administrator or automatic promotion is created.

## Deployment

`server/.env.example` sets `BACKOFFICE_URL=https://backoffice.leamout.com`.
Configure DNS for `backoffice.${DOMAIN}` and the matching Keycloak redirect URI.
The Compose `backoffice` service runs `/leamout/backoffice`; Caddy proxies the
backoffice hostname directly to `backoffice:8082`. No Next.js service or route
is involved. The Docker build generates templates and the local asset bundle.

## Structure

- `cmd/backoffice`: process entry point.
- `internal/runtime/backoffice`: composition, server, routes and operator access.
- `internal/backoffice/auth`: Templ sign-in page and browser authentication routes.
- `internal/backoffice/organizations`: models, repository, handlers, routes,
  validation, pages and feature components.
- `internal/backoffice/agents`: agent models, repository, handlers, routes,
  validation, pages and feature components.
- `internal/backoffice/components`: shared shell and reusable UI.
- `internal/backoffice/assets`: embedded local CSS, JavaScript and favicon.
- `internal/database/queries/organizations.sql`: organization queries, including
  explicit cross-organization backoffice read projections.
- `internal/database/queries/voice_agents.sql`: agent queries, including
  explicit cross-organization backoffice read projections.
- `internal/database/sqlc`: generated query implementations.

Add a service only for real operator orchestration; future mutations should use
the owning domain rules and record operator audit events. Resource counts include
configured agents and trunks regardless of status. Subscription state is the
stored record, not an effective entitlement calculation. Offset pagination uses
a deterministic creation-time/ID order; concurrent insertions may shift pages.

Templ output and generated Tailwind CSS remain untracked. Generate templates
before Go checks and regenerate SQLC after query changes.
