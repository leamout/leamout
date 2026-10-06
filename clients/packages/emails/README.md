# Leamout email templates

React Email owns the HTML layout for OTP and organization invitation emails.
Go owns runtime data validation, subject lines, plain-text rendering, durable
queueing, and AWS SES delivery. This package runs during development and export;
it does not run inside the server or worker container.

From `clients/`:

```sh
bun install --frozen-lockfile
bun run emails:dev
bun run emails:export
bun run emails:check
```

The preview server runs on port 3002 with synthetic `PreviewProps`. The exporter
passes explicit placeholder values instead of preview data, renders complete
HTML documents, and replaces those placeholders with trusted Go template actions.
This replacement happens after HTML formatting so actions containing quotes and
URLs remain intact. Recipient data is inserted only by Go's `html/template` at
runtime, preserving contextual escaping.

Commit the React sources, `clients/bun.lock`, and the generated HTML in
`server/internal/platform/email/templates/` together. Do not edit generated HTML
by hand. CI checks that the exported output matches the committed files and runs
the Go renderer tests. Plain-text `.txt` templates stay maintained in Go.

Use inline styles and React Email components rather than the web application's
UI components, which target browsers. The shared layout lives in
`emails/_components/layout.tsx`. No external images or fonts are required.

When adding an email, define its React component, preview data, export mapping,
Go template data validation and subject, and a plain-text version. Wire its
business operation to the existing transactional queue separately. Invitation
rendering is supported here; this change does not implement invitation endpoints.
