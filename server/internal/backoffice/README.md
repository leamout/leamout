# Leamout Backoffice

The Backoffice is a separate operator-facing HTTP runtime built with Go, Templ, HTMX, TailwindCSS, DaisyUI, and Hyperscript.

Build the local UI assets from `server/internal/backoffice`:

```bash
npm ci
npm run build
```

Generate Templ Go source locally when developing:

```bash
go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate
```

Generated `*_templ.go` files and `assets/static/css/tailwindcss.css` are build artifacts and are not committed.

Then start the Backoffice from `server` (when `DATABASE_URL` points to a
reachable migrated PostgreSQL database):

```bash
go run ./cmd/backoffice
```

The runtime listens directly on `http://127.0.0.1:8081`.

## Complete local browser workflow

The simplest workflow uses the repository Compose stack. From the repository
root:

```bash
cp .env.example .env
# Replace the required placeholder secrets in .env, then:
make up
make migrate
```

`make up` builds and starts the normal API on port 8080 and the Backoffice on
port 8081. The migration command is safe to repeat. For UI-only iteration,
regenerate and rebuild with:

```bash
cd server
go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate
npm ci --prefix ./internal/backoffice --no-audit --no-fund
npm run --prefix ./internal/backoffice build
cd ..
docker compose --env-file .env -f deploy/compose.yaml up -d --build backoffice
```

Promote an existing local account using its email address:

```bash
docker compose --env-file .env -f deploy/compose.yaml exec postgres \
  psql -U leamout -d leamout -c \
  "UPDATE users SET is_platform_admin = TRUE, updated_at = NOW() WHERE lower(email) = lower('admin@example.test') RETURNING id, email, is_platform_admin;"
```

Authenticate through the **normal API**, not the Backoffice. Open
`http://localhost:8080/healthz`, open the browser developer console, and run
the following (substitute the existing account credentials):

```javascript
const started = await fetch("/auth/start", {
  method: "POST",
  credentials: "include",
  headers: { "content-type": "application/json" },
  body: JSON.stringify({ email: "admin@example.test" }),
}).then((response) => response.json());

await fetch("/auth/password/login", {
  method: "POST",
  credentials: "include",
  headers: { "content-type": "application/json" },
  body: JSON.stringify({
    transaction_id: started.data.transaction_id,
    password: "your-password",
  }),
}).then((response) => response.json());
```

Then open **http://localhost:8081/**. Use `localhost` consistently for both
ports: the normal `leamout-session` cookie is host-scoped (cookies are shared
between ports), HTTP-only, Secure, and SameSite=Lax. Production deployments
that use sibling hostnames must set `LEAMOUT_SESSION_COOKIE_DOMAIN` to their
shared parent domain and terminate both runtimes over HTTPS.

The protected request chain is the normal session resolver, the
`users.is_platform_admin` authorization check, and same-origin protection for
unsafe Backoffice requests. The Backoffice does not issue credentials or own a
login endpoint.

## Structure

- `assets` owns the embedded static asset tree.
- `components` owns reusable Templ UI primitives only.
- Feature packages own Backoffice-specific models, cross-tenant repositories, handlers, routes, validation where needed, pages, and feature-specific components.
- Backoffice repositories are operator-facing cross-tenant read projections. They do not replace tenant-scoped domain repositories.
- Backoffice `service.go` files should only be introduced when a feature needs real operator-specific orchestration or mutations.
- `internal/runtime/backoffice/modules.go` owns Backoffice feature composition.
- `internal/runtime/backoffice/routes.go` only mounts feature routes and static/runtime endpoints.
- `internal/runtime/backoffice/server.go` owns the assembled Backoffice HTTP server.

The canonical feature package shape is:

```text
feature/
├── model.go
├── repository.go
├── handler.go
├── routes.go
├── validation.go      # only when the feature has query/form validation
├── pages.templ
└── components.templ
```

Current feature modules are `users`, `organizations`, `calls`, `numbers`,
`trunks`, `carrierconnections`, `providers`, and `commercial`. Users,
organizations, and calls currently have detail pages; the remaining modules
provide cross-tenant list projections while their safe detail projections are
built out.

## Component conventions

Shared components should stay generic and DaisyUI-oriented:

- Use `Class` for DaisyUI and Tailwind variants.
- Use `templ.Attributes` for HTML, HTMX, and Hyperscript attributes.
- Use Templ children or `templ.Component` fields for arbitrary markup.
- Keep feature routes and domain-specific behavior out of shared components.
- Do not use inline JavaScript; the Backoffice CSP only allows scripts loaded from the local origin.
