# Leamout website

The marketing layout provides the shared header, footer, and main landmark. Pages inside `(marketing)` render their content within that landmark.

Run `bun install` from `clients`, then `bun run dev` from this directory. The website runs on port 3000 and links to the console on port 3001 in development. Production console links default to `https://console.leamout.com`.

For another console hostname, set `CONSOLE_URL` in this app's `.env.local` before building. Marketing pages are prerendered, so changing the URL requires a rebuild. Documentation links use the existing `/docs` redirect.

## Marketing content

The homepage uses root metadata. Supporting pages define their own title, description, and canonical URL. Pricing in `lib/plans.ts` mirrors the seed catalog in `server/migrations/033_create_plans.sql`; update both when the catalog changes. Prices are monthly USD platform prices; carrier and AI charges remain with customers' providers.

Before public launch:

- Confirm the private enterprise contact and vulnerability-reporting channels. The contact page currently offers documentation and public GitHub issues only.
- Supply approved privacy, terms, and cookie documents, including the responsible legal entity and contact details. These pages currently explain their pending status and use `noIndex`.
- Supply the verified subprocessor register. This page also uses `noIndex` until published.
- Review deployed analytics and cookie behavior before finalizing the cookie policy.

No certifications, customer testimonials, uptime commitments, or provider-processing locations are assumed in this copy.
