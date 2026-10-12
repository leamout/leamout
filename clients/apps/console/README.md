# Leamout console

Authentication uses React Hook Form and Zod for form validation, TanStack Query for API mutations and session queries, and a small AuthFlowProvider for progress across auth pages.

Run `bun install` from `clients`, then `bun run dev` from `clients/apps/console`. The console listens on http://localhost:3001. Its Next.js rewrite forwards `/api/v1/:path*` to the Go API using `BACKEND_URL` (default http://localhost:8080); set it in the console's `.env.local` when the Go API runs elsewhere. Set `BACKEND_URL` before building; production rewrites are recorded at build time. Session tokens remain in HttpOnly cookies.

For local Go development, set `APP_ENV=development` and `FRONTEND_URL=http://localhost:3001` in `server/.env`. The backend must be running with its dependencies and SES worker for email codes. In production, serve the console over HTTPS and configure the backend `DOMAIN` so its session cookie covers the console hostname (or leave it empty for a host-only cookie). Do not point BACKEND_URL at the console's own `/api` proxy.

## Authentication

- Login starts a transaction, then accepts a password or sends an email code.
- Signup sends and verifies a code before enrolling a password with the new HttpOnly session.
- Recovery starts with an email and collects a code, then submits the code and replacement password together to `/auth/password/reset`. The recovery code is verified by the backend at that final step. Refreshing the reset page requires entering the code again.
- The transaction ID, email, intent, stage, and local expiry/cooldown hints are stored in sessionStorage. Passwords and codes are never persisted. The backend enforces all transaction and session validity.
- Console and onboarding layouts check `/users/me` before displaying their content. API authorization remains enforced by Go. Network errors show retry controls; an expired session returns to login.
- Logout revokes only the current session, clears the cookie and private query cache, and returns to login.

Next.js rewrites forward API requests and cookies. Forms call endpoints directly through the shared JSON request helper. Go rejects untrusted browser origins on writes, requires JSON for auth POST requests, and marks API responses as non-cacheable. `FRONTEND_URL` is the trusted browser origin for writes. The Go auth limiter currently counts the proxy peer IP, so its existing 30 requests/minute budget is shared by browsers using that proxy. Forwarded client IP headers are not accepted in this integration.

Organization selection and other resource pages remain previews and are not integrated in this PR. Successful authentication opens the console overview.

## Console route structure

| Resource | Routes |
| --- | --- |
| Agents | `/agents`, `/agents/new`, `/agents/[agentId]` |
| Calls | `/calls`, `/calls/[callId]` |
| Existing BYOC numbers | `/phone-numbers`, `/phone-numbers/new`, `/phone-numbers/[numberId]` |
| SIP trunks | `/sip-trunks`, `/sip-trunks/new`, `/sip-trunks/[trunkId]` |
| AI providers | `/ai-providers`, `/ai-providers/[providerId]` |
| Tools | `/tools`, `/tools/new`, `/tools/[toolId]` |
| Webhooks | `/webhooks`, `/webhooks/new`, `/webhooks/[webhookId]` |
| Settings | `/settings/profile`, `/settings/security`, `/settings/members`, `/settings/organization`, `/settings/api-keys`, `/settings/storage` |

New resource routes are title and description placeholders pending UI design. Agent configuration will use tabs within agent details. Call diagnostics stay within call details; webhook delivery history stays within endpoint details.

`/sip-trunks` replaces `/connections`. Number purchasing and provisioning, billing pages, and deployment operations pages are deferred until their scope is defined.
