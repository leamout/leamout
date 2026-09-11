# Leamout Agent Instructions

This file is the repository-wide engineering contract for AI coding agents working on Leamout.

## Product

Leamout is a programmable communications control plane for building and operating voice, messaging, numbering, routing, and carrier-connected telecom products.

Leamout must preserve these product principles:

- BYOC is first-class. Customers must not be forced onto Leamout-managed carrier connectivity.
- Self-hosted and Leamout Cloud should expose the same communications model.
- Runtime placement and connectivity ownership are independent choices.
- Carrier and payment providers are adapters, not sources of Leamout domain truth.
- PostgreSQL is the authoritative durable state store.
- Redis is ephemeral coordination/cache state and must never become the monetary source of truth.
- NATS is used for durable asynchronous workflows, not as the authoritative database.
- Telecom history and monetary history must remain durable and auditable.

## Roadmap guardrails

The product roadmap is:

0. Control-plane primitives
1. Self-Hosted + BYOC
2. Self-Hosted + Managed Carrier
3. Leamout Cloud + BYOC
4. Leamout Cloud + Managed Carrier
5. Multi-carrier orchestration
6. Number provisioning + lifecycle
7. Messaging + realtime media + AI
8. Direct carrier connectivity / full CPaaS

Do not introduce later-stage architecture merely because it may be useful someday. Build from the current stable primitives and implement only what the active task or current roadmap stage requires.

Before changing roadmap behavior, read the relevant documents under `docs/`, especially the current voice, BYOC, managed-voice, deployment, and Commercial documents.

If documentation conflicts with current code or a newer explicit architecture decision, do not silently guess. Identify the conflict. Prefer the current implementation plus the newest explicit design decision, and update stale documentation when the task makes that appropriate.

## Repository boundaries

Major server boundaries are:

- `server/internal/identity` — users, authentication, sessions, account identity.
- `server/internal/tenancy` — organizations and tenant ownership.
- `server/internal/commercial` — catalog, billing, access, usage, prepaid value.
- `server/internal/telecom` — communications resources and telecom-domain behavior.
- `server/internal/integrations` — external provider/client adapters.
- `server/internal/runtime` — application composition, transport, middleware, runtime processes.
- `server/internal/backoffice` — operator-facing administrative views and actions.
- `server/internal/database` — SQLC-generated database access and query definitions.

Do not move provider-specific behavior into core Commercial or Telecom domains. Core domains should depend on narrow provider-neutral interfaces.

## Commercial architecture

Commercial is organized as:

```text
Catalog
  products / plans / prices / meters

Billing
  checkout / payments

Access
  subscriptions / licenses / entitlements

Usage
  usage_events

Prepaid
  wallets / wallet ledger / reservations
```

Commercial invariants:

- Leamout uses subscription/license plus prepaid collection. Do not introduce postpaid usage credit without an explicit product decision.
- Checkout owns commercial purchase intent and fulfillment.
- Payments owns provider-independent money collection, payment state, provider access, and provider-event reconciliation.
- A Payment must not directly grant subscriptions, licenses, entitlements, or wallet value.
- A successful payment Settlement is interpreted by Checkout.
- Wallet ledger entries are immutable. Corrections, refunds, and chargebacks are compensating entries.
- Managed-provider obligations require prepaid authorization before Leamout incurs upstream cost.
- Customer-facing prices and upstream wholesale/provider cost are separate concepts.
- `prices` is the customer-facing pricing primitive. Do not resurrect removed parallel customer-rate models without an explicit architecture decision.
- Recording a usage event does not automatically make it billable.
- Money must never be mixed across currencies.

## Telecom architecture

- OpenSIPS owns SIP signaling, authentication boundaries, and SIP routing decisions.
- RTPengine owns media relay/NAT/media-policy enforcement.
- FreeSWITCH is an internal application-media worker for programmable media behavior.
- Coturn provides STUN/TURN relay for browser/WebRTC clients.
- Carrier-specific behavior belongs behind adapters.
- A requested BYOC route must never silently fall back to managed routing.
- Managed routing must never silently consume a customer's BYOC trunk.
- Managed-number provider and managed-termination provider are allowed to differ.
- Do not add multi-carrier/LCR behavior until the single-carrier managed/BYOC contracts and acceptance gates remain reliable.

## Persistence

- Application persistence in Commercial must go through SQLC-generated queries.
- Put SQL in `server/internal/database/queries/*.sql`.
- Do not embed ad-hoc SQL strings in Commercial repositories.
- Do not manually edit generated `server/internal/database/sqlc/*.go` files unless the task is explicitly about generated output; change the source SQL and regenerate instead.
- Preserve organization ownership in SQL constraints and query predicates where practical. Do not rely only on middleware for tenant isolation.
- Preserve idempotency for externally retried operations, provider webhooks, usage ingestion, financial mutations, and reconciliation jobs.
- Treat migrations as immutable once released/applied. Only rewrite migration history when the repository explicitly documents that the migration series is still pre-release.

## Go structure

Use the repository's established module convention. Add a file only when the module owns that responsibility:

- `model.go` — domain models, states, errors, commands, inputs, outputs.
- `repository.go` — durable persistence.
- `service.go` — business rules and orchestration.
- `validation.go` — reusable validation/normalization.
- `handler.go` — HTTP transport.
- `routes.go` — route registration.
- `consumer.go` — asynchronous events entering a module.
- `publisher.go` — asynchronous events leaving a module.
- `jobs.go` — scheduled/background work.

Do not create empty scaffold files for future features.

Prefer narrow interfaces at domain boundaries. Keep repositories boring. Keep HTTP concerns out of domain services. Keep provider SDK/request shapes out of core models when a provider-neutral model is sufficient.

## Generated artifacts

- SQLC output must be reproducible from query sources.
- Templ-generated Go output must be reproducible from `.templ` sources.
- Do not fix generated files by hand when the generator source is wrong.
- If a source change requires regeneration, regenerate all affected artifacts and include them in the same PR.

## Testing and CI

Do not weaken tests to make CI pass.

Acceptance tests are product contracts. When an acceptance test fails because implementation behavior is wrong, fix the implementation rather than removing assertions, replacing real behavior with mocks, bypassing readiness, or increasing arbitrary sleeps.

For server Go changes, run the relevant checks used by CI, including:

```text
gofmt
go test ./...
golangci-lint
```

Also run targeted package tests while iterating.

When SQL/query sources change, regenerate SQLC and verify that regeneration is clean. When Templ sources change, regenerate Templ output. Check the repository workflows/Makefile for the exact current commands rather than inventing a different tool version.

Security-sensitive or telecom-runtime changes should preserve the relevant acceptance gates, including BYOC, Voice, WebRTC, managed carrier, SIP edge, and graceful-drain behavior where applicable.

## Change discipline

- Start from current `main` unless the task explicitly targets another branch/stack.
- Read the current implementation before proposing a replacement architecture.
- Keep PRs focused around one coherent boundary or vertical slice.
- Avoid compatibility aliases unless a real staged migration requires them.
- Do not leave dead transitional code after all callers are migrated.
- Do not introduce generic abstractions before there are concrete use cases that need them.
- Do not expose provider credentials, provider resource IDs, password material, session tokens, or other secrets through customer APIs or logs.
- Preserve existing public behavior unless the task explicitly changes the contract.
- Update docs when the architecture or public contract changes.

## Definition of done

A change is complete when:

1. the domain boundary is clear;
2. persistence and tenant ownership remain correct;
3. retries/idempotency are handled where the operation can repeat;
4. tests cover the new behavior and important failure cases;
5. formatting, tests, linting, and required generation pass;
6. relevant acceptance guarantees are not weakened;
7. documentation reflects material architecture or API changes;
8. the PR explains the invariant being introduced or preserved.
