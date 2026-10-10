# Leamout console

Auth pages live in `app/(auth)`, each using a dedicated form in `components/auth`. React Hook Form and Zod handle form validation. The shared AuthForm handles presentation; each dedicated form handles its own preview submission feedback.

Onboarding pages live in `app/(onboarding)`, with dedicated components in `components/onboarding`. `/create-organization` validates an organization name. `/organizations` displays an empty state by default; the selector component also supports an organization list for UI previews.

Console pages live in `app/(console)` with the sidebar.

These pages are UI previews. No API requests, transaction storage, sessions, or backend guards are integrated. Submitting forms displays a preview notice without creating accounts or organizations. Links allow reviewing the page designs.

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
