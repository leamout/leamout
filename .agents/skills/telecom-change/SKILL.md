---
name: telecom-change
description: Implement or review Leamout telecom changes involving calls, SIP, carriers, trunks, routing, numbers, managed voice, WebRTC, media, CDRs, or provider adapters.
---

# Work in Leamout Telecom

Read `AGENTS.md` and the relevant telecom roadmap/documentation before editing.

## Core invariants

- BYOC remains first-class.
- Runtime placement and carrier ownership are independent.
- An explicit BYOC request must never silently fall back to managed connectivity.
- Managed routing must never consume another organization's BYOC resources.
- Carrier/provider-specific behavior belongs behind adapters.
- OpenSIPS owns SIP signaling/routing policy enforcement at the edge.
- RTPengine owns media relay/NAT/WebRTC media boundary.
- FreeSWITCH is an internal programmable media worker, not the tenant/routing authority.
- Durable call history and tenant attribution must survive restart/reconciliation.
- Provider credentials and provider resource IDs remain internal unless an API explicitly exposes safe projections.

## Managed telecom

Treat managed-provider work as a Leamout financial obligation. Preserve fail-closed commercial authorization and do not bypass prepaid/risk controls. Keep wholesale provider cost separate from customer pricing.

## Verification

Prefer real acceptance topology over mocks for SIP/media/routing behavior. When changing managed behavior, verify BYOC behavior still passes, and vice versa.