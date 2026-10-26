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
dependencies with a frozen lockfile, runs Biome and TypeScript checks, and
builds the marketing app.

`app/opengraph-image.tsx` generates a 1200×630 branded image at build/request
time. `lib/metadata.ts` uses this route for Open Graph and Twitter previews.

## Content and contact

The blog uses MDX files in `content/` with validated metadata. Drafts are
excluded from the static article routes and sitemap.

Pricing in `lib/plans.ts` currently mirrors
`server/migrations/033_create_plans.sql`; review both whenever the
subscription catalog changes. Public pricing covers the managed runtime.
Carrier and AI provider charges are separate where applicable.

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
