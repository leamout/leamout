# Leamout

**Leamout** is a programmable communications control plane for building and operating voice, messaging, numbering, routing, and carrier-connected telecom products.

Leamout starts with self-hosted programmable voice and BYOC, then grows toward Leamout Carrier, Leamout Cloud runtimes, Cloud-managed connectivity, multi-carrier orchestration, number provisioning, messaging, realtime media, and AI communications.

The goal is to give applications a stable communications API without forcing customers to give up control of their telecom infrastructure or carrier relationships.

## Platform model

Leamout separates the customer-facing control plane from communications resources, runtime placement, carrier connectivity, and telecom execution.

`console.leamout.com` is the planned shared management surface for both self-hosted and Leamout Cloud customers.

```text
                            console.leamout.com
                                   │
                                   ▼
                         Leamout Control Plane
                                   │
             ┌─────────────────────┼─────────────────────┐
             │                     │                     │
          Identity             Commercial              Fleet
             │                     │                     │
             └─────────────────────┼─────────────────────┘
                                   │
                             Organizations
                                   │
                    Communications Resources
                                   │
          ┌────────────────────────┼────────────────────────┐
          │                        │                        │
        Voice                    Messaging               Numbers
          │                        │                        │
          └────────────────────────┼────────────────────────┘
                                   │
                              Applications
                                   │
                    ┌──────────────┼──────────────┐
                    │              │              │
                  Calls         Messages       Number Flows
                    │              │              │
                    └──────────────┼──────────────┘
                                   │
                           Routing / Policy Engine
                                   │
                                   ▼
                                Runtime
                                   │
                 ┌─────────────────┴─────────────────┐
                 │                                   │
            Self-Hosted                         Leamout Cloud
                 │                                   │
                BYOC                          ┌───────┴───────┐
                 │                            │               │
        customer selects                    BYOC          Managed
        the carrier                          │               │
                 │                     customer selects   Leamout supplies
        ┌────────┼────────┐              the carrier      connectivity
        │        │        │
     carrier  carrier  Leamout
       A        B      Carrier
```

This produces three product delivery modes:

| Runtime | Connectivity | Delivery mode |
| --- | --- | --- |
| Self-Hosted | Customer-selected carrier | **Self-Hosted + BYOC** |
| Leamout Cloud | Customer-selected carrier | **Leamout Cloud + BYOC** |
| Leamout Cloud | Leamout-supplied connectivity | **Leamout Cloud + Managed** |

There is no separate **Self-Hosted + Managed Carrier** mode. A self-hosted customer can select Leamout Carrier just as they can select another supported carrier. In that relationship Leamout is acting as the telecom provider, while the runtime remains Self-Hosted + BYOC.

The resource boundary follows ownership rather than provider brand:

```text
organization-scoped carrier connection
        = customer-selected / BYOC

platform-scoped carrier connection
        = internal Leamout upstream / Cloud Managed
```

An organization-scoped carrier connection may use `provider = leamout`; it is still BYOC. Physical trunk endpoints inherit that ownership from the carrier connection and do not need an independent BYOC/managed type.

BYOC remains a first-class model. Carrier-specific behavior belongs behind adapters so the same control plane can work with customer-selected carriers, Leamout Carrier, Cloud-managed connectivity, and multiple markets.

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

- **One control plane, two runtime distributions.** Self-Hosted and Leamout Cloud should expose the same communications resource model without becoming forks.
- **Self-Hosted is BYOC.** A self-hosted customer selects and configures the carrier. Leamout Carrier is one possible carrier, not a separate self-hosted mode.
- **Cloud supports BYOC and Managed.** Cloud customers may select their own carrier or use connectivity supplied and operated by Leamout.
- **Ownership beats provider brand.** Organization-scoped carrier connections are BYOC even when the provider is Leamout; platform-scoped carrier connections are internal Cloud-managed upstream resources.
- **BYOC stays open.** Customers should not be forced onto Leamout carrier connectivity.
- **The control plane owns business logic.** Routing, policy, usage, rating, billing, provisioning, and events should not be delegated to upstream carriers.
- **Carrier integrations stay behind adapters.** Market-specific connectivity should not leak into core platform APIs.
- **Telecom history is durable.** Historical calls, legs, events, and number assignments must not disappear through cascading configuration deletes.
- **Organization boundaries are database-enforced where practical.** Cross-project telecom references should be rejected by relational constraints, not merely application convention.
- **Build from stable primitives.** Higher-level CPaaS products should depend on reliable voice, routing, events, usage, and billing foundations.

The threat boundaries and production deployment requirements for Cloud,
self-hosted, and telecom workloads are documented in the
[security model](docs/security.md).

## Roadmap

```text
0. Control-plane primitives
        ↓
1. Self-Hosted + BYOC
        ↓
2. Leamout Carrier as a BYOC provider
        ↓
3. Leamout Cloud + BYOC
        ↓
4. Leamout Cloud + Managed
        ↓
5. Multi-carrier orchestration
        ↓
6. Number provisioning + lifecycle
        ↓
7. Messaging + realtime media + AI
        ↓
8. Direct carrier connectivity / full CPaaS
```
