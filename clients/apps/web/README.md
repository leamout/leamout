# Leamout website

The Next.js marketing website lives in `clients/apps/web`. Marketing routes use
the shared layout, header, footer, and main landmark in `app/(marketing)`.
Sections live under their page-specific directories in `components/`, and the
shared navigation lives in `components/marketing/`.

## Local development

From `clients`, install dependencies with `bun install --frozen-lockfile`.
From `clients/apps/web`, use:

```bash
bun run dev
bun run lint
bun run typecheck
bun run build
```

The website runs on port 3000 and links to the console on port 3001 during
development. Production console links default to
`https://console.leamout.com`. To override this, set `CONSOLE_URL` in the web
app's environment before building; marketing pages are prerendered.

The site redirects `/docs` to `https://docs.leamout.com`, and `/status` to
`https://status.leamout.com`.

## Launch and SEO checks

The separate [web workflow](../../../.github/workflows/web.yml) installs
dependencies with a frozen lockfile, runs Biome and TypeScript checks,
and builds the marketing app.

`app/opengraph-image.tsx` generates a 1200×630 branded image at build/request
time. `lib/metadata.ts` uses this route for Open Graph and Twitter previews.

## Content and contact

The blog uses MDX files in `content/` with validated metadata. Drafts are
excluded from the static article routes and sitemap.

Public fixed-price plans in `lib/plans.ts` mirror
`server/migrations/033_create_plans.sql`. The backend also seeds a custom
Enterprise plan, which is not displayed as a fixed-price card. Review prices,
limits, and entitlements in both catalogs whenever subscription offerings
change. Carrier and AI provider charges are separate where applicable.

The contact page always provides a public GitHub issue tracker for non-sensitive
product feedback and a security disclosure email as currently configured.
For general private inquiries, the page links directly to
`hello@leamout.com`. Verify mailbox delivery before public launch. Never use the
public issue tracker to request secrets, customer call data, or recordings.

## Before accepting public paid customers

- Publish approved privacy, service terms, and cookie documents, including the
  responsible legal entity, terms of processing, and applicable rights.
- Publish the verified subprocessor register. Until approval, those routes
  remain `noIndex` and explain that the documents are pending.
- Confirm the deployed Vercel Analytics and Speed Insights behavior and any
  consent obligations.
- Verify general and security mailboxes can receive inquiries.
- Verify the production console signup flow, SIP onboarding, and service
  availability before making unconditional production-readiness claims.

The marketing site does not claim external certifications, customer
testimonials, uptime guarantees, or audited provider-processing locations.

## Product references and onboarding

The homepage uses a static, conceptual SIP-to-runtime-to-AI diagram and
representative call event names. These visuals are explanatory references, not
live activity, authenticated Console screenshots, customer records, or
evidence of production traffic. The marketing shell and homepage use Leamout's
light palette: off-white surfaces, high-contrast typography, and orange
accents.

Keep product capabilities aligned with the current Console implementation:

- `clients/apps/console/components/agents/agent-details.tsx` defines the
  agent's available tabs.
- `clients/apps/console/components/agents/agent-form.tsx` defines the
  configuration fields.
- `clients/apps/console/components/agents/agent-readiness.tsx` defines the
  readiness report.
- `clients/apps/console/components/calls/call-details.tsx` defines persisted
  call diagnostics and API limitations.
- `clients/apps/console/components/overview/overview-dashboard.tsx` defines
  organization setup steps.

The four onboarding steps link directly to published docs in
`leamout/docs`: AI integrations, SIP trunks, Voice Agents, and inbound calls.
The API readiness example is documented in the official quickstart. The
marketing website never requests provider secrets or embeds organization data.

For a future screenshot-based showcase, capture approved screens from a
dedicated demo organization with explicitly sanitized data and verify that
features are available in the deployed environment. Never label reconstructed
interface references as live screenshots.

## Shared infrastructure

The marketing website does not use the console's toast or tooltip providers.
Its root layout stays a server component without a site-wide client wrapper.
Vercel Analytics and Speed Insights remain mounted in `app/layout.tsx`;
review deployment and consent requirements before public launch.

The shared `@leamout/ui` package owns the Base UI, Hugeicons, `cn` and
Recharts dependencies used by its components, so the web app no longer
redeclares them. Unused `motion` is removed. Next.js, MDX, schema and Vercel
dependencies remain for existing functionality. Keep `clients/bun.lock` in
sync with any workspace manifest changes; CI uses `--frozen-lockfile`.

## Marketing page responsibilities

The homepage introduces the managed runtime, explains the overall connection,
shows source-aligned call operations, and links developers to onboarding.
The Platform page explains SIP signaling, realtime media, agent orchestration,
and the call lifecycle. The Integrations page covers the setup of existing
SIP trunks, AI provider credentials, and application tools. Avoid repeating
the same three-layer connection diagram across these pages.

The interface references are still not screenshots of a live Console or
evidence of customer traffic. Replace them with sanitized screenshots only
after verifying the actual deployed demo environment.
