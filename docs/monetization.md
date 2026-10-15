# Leamout commercial model

Leamout is the sole supported product: a managed carrier-grade runtime
and control plane for autonomous voice agents. Leamout runs the infrastructure;
customers manage their organizations, agents, carrier connections and AI
provider accounts through the Leamout console and API.

## Customer workflow

```text
Create account and organization
    ↓
Select subscription plan
    ↓
Connect customer-owned SIP trunk and phone numbers (BYOC)
    ↓
Configure organization AI providers (BYOA)
    ↓
Create and activate Voice Agent
    ↓
Handle calls, inspect diagnostics and usage
```

## Revenue and responsibility

Leamout subscriptions fund the managed runtime, control plane,
observability and platform capacity. Carrier charges are billed by the
customer's carrier. AI providers are billed directly to the customer when
using BYOA. There is no managed carrier resale, telecom wallet, telecom
usage rating, or prepaid telecom balance.

The current application and website plan catalog is seeded from
`server/migrations/033_create_plans.sql`. Do not publish different prices
or capacity limits without changing and validating that catalog and the
Polar product mappings together.

| Plan | Current seed price (USD/month) |
| --- | ---: |
| Free | $0 |
| Developer | $79 |
| Pro | $299 |
| Scale | $799 |

These are current repository seed values, **not** a decision to change or
finalize pricing. Plan limits cover concurrent calls, monthly runtime
allowance, number of Voice Agents and SIP trunks, and retention.

## Plan enforcement

The API and asynchronous worker must apply the same organization entitlement
and runtime-usage checks to every new call/agent session. Missing plan data
must not grant unlimited capacity. Once the runtime allowance is exhausted,
new sessions are rejected; active calls can finish. Subscription changes,
payment-grace periods, and allowance reset use the durable subscription
lifecycle and Polar reconciliation logic.

## AI provider ownership

Voice Agent STT, LLM, TTS and realtime providers bind organization-owned
provider integrations. Secrets are encrypted and must not be copied into
immutable session snapshots. Any platform-supplied AI feature and usage
billing must be defined explicitly before customers are charged for it.

## Recording storage

Managed recording storage is organization-isolated. Customers may
optionally configure a tenant-owned S3-compatible recording destination
(BYOS). This does not change Leamout's ownership of the Managed infrastructure.
See [storage.md](storage.md) for the security model.

## Operations

Leamout deploys and operates its own SIP, media, API, worker, database,
coordination, storage and observability services. The repository's Docker
Compose and acceptance scripts are for Leamout's internal operations and
testing, **not** a customer-supported Self-Hosted product.
