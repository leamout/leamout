# Leamout

**Leamout** is a programmable communications control plane for building and operating voice, messaging, numbering, routing, and carrier-connected telecom products.

Leamout starts with self-hosted programmable voice and BYOC, then grows toward managed deployments, multi-carrier orchestration, number provisioning, direct carrier connectivity, messaging, realtime media, and AI communications.

The goal is to give applications a stable communications API without forcing customers to give up control of their telecom infrastructure or carrier relationships.

## Platform model

```text
                         Leamout
                            │
             ┌──────────────┼──────────────┐
             │              │              │
           Voice           SMS           Numbers
             │              │              │
             └──────────────┼──────────────┘
                            │
                     Routing engine
                            │
          ┌─────────────────┼─────────────────┐
          │                 │                 │
     Customer BYOC       Leamout       Other carrier
          │                 │                 │
          └─────────────────┼─────────────────┘
                            │
                       Telecom networks
```

BYOC remains a first-class model. Carrier-specific behavior belongs behind adapters so the same control plane can work with customer-owned carriers, Leamout-managed connectivity, and multiple markets.

## Current implementation

The repository currently contains the foundations of the Leamout control plane:

- users, accounts, memberships, projects, and account tokens
- project-scoped SIP domains and endpoints
- rotation-safe SIP digest credentials
- carrier providers, project carrier connections, trunks, and multiple physical SIP endpoints
- inbound and outbound routing primitives
- phone-number inventory and project assignments
- calls, call legs, and durable call events
- transactional outbox primitives for asynchronous workflows
- SQLC-generated database access and Atlas-managed schema workflow
- a deployable development telecom stack around OpenSIPS, RTPengine, FreeSWITCH, Coturn, PostgreSQL, Redis, and NATS
- browser/WebRTC acceptance coverage that forces TURN relay, registers over WSS, and verifies real RTP media through RTPengine and FreeSWITCH

Some capabilities described in the documentation are intentionally **roadmap design**, not current product behavior. In particular, full billing/rating, SMPP messaging, realtime AI media, production HA, and hosted carrier products are not presented as complete.

Implementation progress for programmable call control is tracked in the
[call orchestration roadmap](docs/call-orchestration-roadmap.md). BYOC, secure
WebRTC, and carrier-connectivity progress is tracked in the
[BYOC roadmap](docs/byoc-roadmap.md).

The target customer installation, activation, upgrade, backup, and operator CLI
contract is documented in the
[self-hosted installation guide](docs/deploy/self-hosted-installation.md).

## Architecture

Leamout separates control-plane business logic from telecom data-plane components.

```text
Applications / API clients
          │
          ▼
   Leamout control plane
          │
          ├── PostgreSQL ── source of truth
          ├── Redis ─────── ephemeral state / coordination
          └── NATS ──────── durable asynchronous workflows
          │
          ▼
       OpenSIPS ◄──── browser SIP over WSS
          │
     ┌────┴────┐
     │         │
RTPengine   FreeSWITCH
     │         │
     └────┬────┘
          │
   SIP carriers / PSTN

Browser media ── TURN/STUN via Coturn ──► RTPengine
```

**OpenSIPS** owns SIP signaling and routing. **RTPengine** handles media relay and NAT/media anchoring, including the WebRTC media boundary. **FreeSWITCH** is used when calls require application media such as IVR, playback, recording, conferencing, or similar programmable media workloads. **Coturn** provides STUN/TURN relay service for browser/WebRTC clients; Leamout issues the short-lived tenant-bound TURN credentials used by those clients.

## Core principles

- **BYOC stays open.** Customers should not be forced onto Leamout carrier connectivity.
- **The control plane owns business logic.** Routing, policy, usage, rating, billing, provisioning, and events should not be delegated to upstream carriers.
- **Carrier integrations stay behind adapters.** Market-specific connectivity should not leak into core platform APIs.
- **Telecom history is durable.** Historical calls, legs, events, and number assignments must not disappear through cascading configuration deletes.
- **Organization boundaries are database-enforced where practical.** Cross-project telecom references should be rejected by relational constraints, not merely application convention.
- **Build from stable primitives.** Higher-level CPaaS products should depend on reliable voice, routing, events, usage, and billing foundations.

## Roadmap

```text
0. Control-plane primitives
        ↓
1. Self-hosted programmable voice + BYOC
        ↓
2. Managed programmable voice
        ↓
3. Multi-carrier orchestration
        ↓
4. Number provisioning + lifecycle
        ↓
5. Direct carrier connectivity
        ↓
6. Messaging + realtime media + AI
        ↓
7. Hosted carrier products / full CPaaS
```
