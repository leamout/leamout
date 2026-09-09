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

Then start the Backoffice from `server`:

```bash
go run ./cmd/backoffice
```

The runtime listens directly on `http://127.0.0.1:8081`.

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

Current feature modules are `organizations`, `calls`, `numbers`, `trunks`, `carrierconnections`, `providers`, and `commercial`.

## Component conventions

Shared components should stay generic and DaisyUI-oriented:

- Use `Class` for DaisyUI and Tailwind variants.
- Use `templ.Attributes` for HTML, HTMX, and Hyperscript attributes.
- Use Templ children or `templ.Component` fields for arbitrary markup.
- Keep feature routes and domain-specific behavior out of shared components.
- Do not use inline JavaScript; the Backoffice CSP only allows scripts loaded from the local origin.
