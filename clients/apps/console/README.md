# Leamout Console

Authentication is provided by Keycloak through the Go backend. Login and registration use browser redirects; no password or verification code is collected by the console. Account profile and security settings open the identity account console.

The existing `/api/v1` rewrite proxies login, the callback, and API calls to the backend. Keycloak tokens remain encrypted on the server. The browser receives only an HttpOnly, host-only session cookie. Logout clears that session and navigates to Keycloak to end SSO.

Set the backend `FRONTEND_URL` to this console's public origin (for local `bun dev`, `http://localhost:3001`). Register exactly `<FRONTEND_URL>/api/v1/auth/callback` in Keycloak. See `docs/keycloak.md` at the repository root for setup.

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
