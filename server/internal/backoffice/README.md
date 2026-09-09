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
- Feature packages such as `dashboard`, `organizations`, and `calls` own their handlers, routes, pages, page-specific components, and Backoffice-only view models or validation when needed.
- Backoffice feature packages reuse existing Leamout repositories for reads and domain services for writes rather than creating a parallel service or persistence layer.
- `internal/runtime/backoffice` owns HTTP runtime concerns and mounts Backoffice feature modules.
