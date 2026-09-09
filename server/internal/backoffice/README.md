# Leamout Backoffice

The Backoffice is a separate operator-facing HTTP runtime built with Go, Templ, HTMX, TailwindCSS, DaisyUI, and Hyperscript.

Build the local UI assets from `server/internal/backoffice`:

```bash
npm ci
npm run build
```

Then start the Backoffice from `server`:

```bash
go run ./cmd/backoffice
```

The runtime listens directly on `http://127.0.0.1:8081`.

Templ components live in `components`. After changing a `.templ` file, regenerate
the checked-in Go source before running the test suite:

```bash
go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate ./internal/backoffice/components
```
