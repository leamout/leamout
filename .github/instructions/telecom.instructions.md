---
applyTo: "server/internal/telecom/**/*.go,docs/**/*voice*.md,docs/byoc-roadmap.md,docs/providers/**/*.md"
---

# Telecom Instructions

Follow `AGENTS.md` first. These rules apply specifically to telecom behavior.

## Core responsibilities

- OpenSIPS owns SIP signaling, ingress authentication boundaries, and SIP routing.
- RTPengine owns media relay, NAT traversal, and media policy enforcement.
- FreeSWITCH is an internal programmable-media worker, not the source of tenant/carrier policy.
- Coturn provides STUN/TURN for WebRTC clients.
- Carrier/provider-specific behavior belongs behind adapters.

## BYOC and managed routing

- BYOC remains first-class.
- Explicit BYOC requests must only use organization-owned BYOC resources.
- A failed requested BYOC route must not fall back to managed connectivity.
- Managed routing must not silently use an organization's BYOC trunk.
- Platform-scoped managed carrier resources are internal and must not leak as another tenant's resources.
- Managed number provider and managed termination provider may differ.
- Do not introduce multi-carrier/LCR behavior before the single-carrier acceptance contracts remain reliable.

## Tenant and identity boundaries

- Resolve organization ownership from authoritative resources; do not infer tenancy from untrusted SIP metadata.
- Strip/rewrite externally supplied Leamout routing metadata at trusted ingress boundaries.
- Preserve organization ownership in database queries and relational constraints where practical.
- Do not expose provider credentials, internal provider resource IDs, or private route metadata through customer APIs.

## Runtime reliability

- Fail closed when authorization, entitlement, quota, or mandatory dependency state is unavailable.
- Preserve health/failover semantics for trunks/endpoints.
- Preserve restart/reconciliation behavior for durable calls and runtime leases.
- Do not replace readiness/reconciliation fixes with arbitrary sleeps.

## Testing

Changes that touch routing, call lifecycle, media, credentials, quotas, managed admission, or carrier behavior should run the relevant targeted unit tests and applicable acceptance gates.

Never weaken BYOC, Voice, WebRTC, managed SIP edge, managed-carrier, or graceful-drain assertions merely to make CI pass.
