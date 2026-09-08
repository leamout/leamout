# Leamout Backoffice

The Backoffice is a separate operator-facing HTTP runtime built with Go, Templ, HTMX, TailwindCSS, DaisyUI, and Hyperscript.

Build the local UI assets from `server/internal/backoffice`:

```bash
npm install
npm run build
```

Then start the Backoffice from `server`:

```bash
go run ./cmd/backoffice
```

The runtime listens directly on `http://127.0.0.1:8081`.
