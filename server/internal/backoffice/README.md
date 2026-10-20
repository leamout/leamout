# Leamout Backoffice

The operator interface is a separate Go HTTP process using Templ, Tailwind,
DaisyUI and locally embedded assets. It listens on port 8082 and serves pages
under `/backoffice`. The first feature is read-only organization inspection:
name/ID search, lifecycle filters, 25-row pagination, resource counts and
recorded subscription state. Deleted and disabled organizations remain visible.

## Local development

Use the same migrated PostgreSQL database, `ENCRYPTION_KEY`, `KEYCLOAK_*` and
`FRONTEND_URL` as the API. `APP_ENV` controls development HTTPS validation.
The backoffice loads `server/.env` but does not require Redis, NATS, FreeSWITCH,
TURN or payment credentials.

From the repository root, with Go from `server/go.mod` and Bun 1.3.14 installed:

```sh
make backoffice-assets backoffice-generate
cd server
go run ./cmd/backoffice
```

Run the console and API normally. Sign in through the console's existing
Keycloak login, then visit `http://localhost:3000/backoffice`. The console
rewrites `/backoffice/*` to `BACKOFFICE_URL` (default `http://localhost:8082`).
Use the same hostname for login and backoffice; the session cookie is host-only.
An unauthenticated request receives 401; a signed-in non-operator receives 403.
No separate login or credentials are issued by this process.

An existing, verified operator account must be explicitly promoted by a trusted
database administrator, using its exact user ID:

```sql
UPDATE users
SET is_platform_admin = TRUE, updated_at = NOW()
WHERE id = '<verified-operator-user-id>' AND disabled_at IS NULL;
```

There is no default operator account or automatic promotion. Removing the flag
or disabling the user denies subsequent requests. Session resolution uses the
normal encrypted Keycloak token refresh and revocation behavior.

## Deployment

The shared server Docker image builds the asset bundle and generates Templ
sources before compiling `/leamout/backoffice`. Compose includes an internal
`backoffice:8082` service without a published host port. Configure the console's
`BACKOFFICE_URL` at build time to a reachable private backoffice address (for a
console on the Compose control network, `http://backoffice:8082`). The console
must have network access to that service. Keep `FRONTEND_URL` equal to the public
console origin. An external reverse proxy can equivalently forward
`/backoffice/*`, preserving that prefix, on the console's HTTPS hostname.

Do not deploy this behind a different public hostname and expect the console's
host-only cookie to work. There is no need to widen its cookie domain.

## Boundaries

- `cmd/backoffice` starts the process.
- `internal/runtime/backoffice` composes concrete dependencies, authentication,
  database-backed operator authorization, origin protection and routes.
- `assets` embeds static files; `components` contains the shared shell and UI.
- `organizations` owns `model.go`, `repository.go`, `handler.go`, `routes.go`,
  `validation.go`, `pages.templ` and `components.templ`.
- Operator SQL is explicit cross-organization read projection SQL under
  `internal/database/queries/backoffice_organizations.sql`.
- Add `service.go` only when operator orchestration is needed. Future mutations
  must use the owning domain rules and record operator audit events.

Only implemented navigation items are shown. Counts include configured agents
and trunks regardless of their status; active memberships exclude disabled
memberships. Subscription status is the stored record, not an entitlement
calculation. Pagination uses a deterministic creation-time/ID order; concurrent
insertions can shift offset-based pages.

Generated `*_templ.go` and Tailwind CSS remain untracked. Generate templates
before Go tests, builds, lint or vulnerability checks. Run `sqlc generate` after
query changes. `TEST_DATABASE_URL` enables the PostgreSQL-backed access and
organization tests, including revoked sessions and cross-organization reads.
