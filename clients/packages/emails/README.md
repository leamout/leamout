# Leamout email templates

React Email owns the HTML layout for authentication, onboarding, membership, security, and operational emails.
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
business operation to the transactional queue in the same database transaction.

Available templates: `otp`, `invitation`, `welcome`, `invitation-accepted`,
`security-alert`, `api-key-expiry`, `sip-trunk-failure`, and `voice-agent-failure`.
The six notification templates have preview data and Go rendering support;
welcome, password-change security alerts, invitations, and invitation-accepted
notifications are connected to business transactions. Operational alerts and
key expiry reminders still require recipient preferences and deduplication. Pass a sanitized failure summary, never raw provider errors,
credentials, transcripts, or API key secrets.

`Request.ExpiresAt` is the delivery deadline required by the queue.
`Data.ExpiresAt` is the expiry shown only by OTP and invitation templates. API key expiry uses `Data.KeyExpiresAt`, and
security/failure notifications use `Data.OccurredAt` for the event timestamp.

Invitation emails currently target `https://DOMAIN/invitations/accept`. The
backend acceptance endpoint is implemented; the frontend acceptance page is
deferred, so the emailed link does not yet provide a complete user flow.
