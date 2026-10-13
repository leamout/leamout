# Leamout

**Leamout is a carrier-grade platform for autonomous voice agents. Run it
yourself or use Leamout Cloud.**

Leamout sits between customer-owned telephony infrastructure and realtime AI
systems. It owns the runtime required to establish, control, process, and
observe autonomous voice calls while customers keep their carrier relationships,
phone numbers, SIP trunks, PBXs, AI provider accounts, and provider billing.

Leamout is not a general-purpose CPaaS, telecom reseller, or managed carrier.

## Architecture

Leamout is organized around three explicit runtime boundaries:

```text
                              Leamout

+------------------------------------------------------------------+
|                         CONTROL PLANE                            |
|                                                                  |
|  Auth  Tenancy  Agents  Routing  Credentials  API  Webhooks      |
|  Observability  Organizations  Voice Agents  Credentials         |
+-------------------------------+----------------------------------+
                                |
                                | configuration and control
                                v
+------------------------------------------------------------------+
|                         AGENT RUNTIME                            |
|                                                                  |
|  Agent sessions  Turn control  Barge-in  Model orchestration     |
|  Tool execution  Conversation state  Human handoff               |
+-------------------------------+----------------------------------+
                                |
                                | realtime audio
                                v
+------------------------------------------------------------------+
|                       TELEPHONY RUNTIME                          |
|                                                                  |
|  OpenSIPS        FreeSWITCH        RTPengine        Coturn        |
|  SIP routing     Call execution    Media boundary   WebRTC TURN   |
+-------------------------------+----------------------------------+
                                |
                                v
                       customer-owned
                         connectivity

                 SIP carrier · PBX · SBC · WebRTC
```

The control plane configures the runtime but does not sit in the live audio hot
path. The Agent Runtime owns per-call conversational execution. The Telephony
Runtime owns SIP signaling, call control, and media integration.

See [docs/architecture.md](docs/architecture.md) for the canonical system
architecture and [docs/media-plane.md](docs/media-plane.md) for realtime media
contracts and runtime status.

## Product boundary

Leamout provides:

- carrier-grade SIP and media infrastructure;
- programmable call control;
- realtime autonomous voice-agent sessions;
- turn detection and interruption handling;
- composable STT, LLM, and TTS orchestration;
- integrated realtime model support;
- secure tool execution;
- routing and human handoff;
- recordings, events, webhooks, and observability;
- Self-Hosted and Leamout Cloud deployment options.

The same voice-agent platform underlies both options. Enterprise is optional
commercial support, security, governance, deployment assistance, and
contractual packaging for Cloud or Self-Hosted; it is not a separate product
architecture.

Leamout does not buy or resell carrier minutes, phone numbers, SMS, WhatsApp
capacity, or managed carrier services.

## BYOC only

SIP connectivity is organization-owned.

A customer connects an existing SIP carrier, PBX, SBC, or supported WebRTC
endpoint through a Leamout SIP trunk and continues using the external accounts
and phone numbers they already control.

```text
Customer carrier / PBX / SBC
             |
             v
       Leamout SIP trunk
             |
             v
          OpenSIPS
             |
             v
        FreeSWITCH
             |
      RTPengine media
             |
             v
   Leamout Media Runtime
             |
             v
    Leamout Agent Runtime
```

A SIP trunk is Leamout's technical connectivity primitive. Its remote endpoints,
authentication, source networks, codecs, and limits are stored in the control
plane database. OpenSIPS validates and routes against that customer-owned trunk
configuration; it is not a Leamout-provided carrier or shared SIP service.

The peer may be a carrier, PBX, SBC, or another SIP platform. Provider identity
is not a separate runtime domain object.

## Agent runtime

The Agent Runtime owns the state and execution of a live autonomous voice
session:

- conversation and turn state;
- interruption and barge-in handling;
- streaming model orchestration;
- tool-call authorization and execution;
- playback cancellation;
- human handoff;
- runtime events and diagnostics.

Latency-sensitive session state stays local to the active runtime whenever
possible. Redis is used for distributed coordination where cross-process state
is actually required.

The media runtime supports two engine models:

- **Composable**: streaming STT, LLM, and TTS providers coordinated by Leamout.
- **Integrated**: an end-to-end realtime voice model.

Current provider integrations include Deepgram, Groq, Cartesia, and OpenAI
Realtime. Every active provider binding must reference a verified,
organization-owned AI integration. Missing integrations block agent readiness
and runtime execution; deployment-global keys are not used as fallback.
Provider packages adapt external protocols into Leamout's provider-neutral
media contracts; providers do not define the product architecture.

## Media path

Live audio stays on the realtime media path.

Audio frames are not published through NATS JetStream and are not persisted to
PostgreSQL.

```text
Caller RTP
    |
    v
RTPengine
    |
    v
FreeSWITCH
    |
    | bidirectional audio stream
    v
Go Media Runtime
    |
    v
Agent Runtime
    |
    +--> composable STT -> LLM -> TTS
    |
    +--> integrated realtime model
```

Immediate actions such as interruption, playback cancellation, and turn
transitions occur inside the runtime. Durable events can be emitted
asynchronously after the runtime has acted.

## Telephony runtime

The telephony layer remains a first-class part of the product.

- **OpenSIPS**: SIP routing, authentication, and policy enforcement using the
  deployment's database-backed trunk configuration.
- **FreeSWITCH**: B2BUA, call application, and media-control runtime.
- **RTPengine**: RTP anchoring and media boundary.
- **Coturn**: STUN/TURN services when WebRTC connectivity requires them.
- **SIP trunks**: organization-owned authentication, source CIDRs, codecs,
  admission limits, and remote gateway endpoints.
- **Phone numbers**: customer-owned voice bindings attached directly to trunks.
- **Calls and recordings**: programmable voice primitives used by
  the Agent Runtime.

## Control plane

The control plane owns durable configuration and management:

- organizations and identities;
- agents and agent configuration;
- SIP trunks and routing policy;
- model and tool credentials;
- API and authorization;
- organization, user, and deployment administration;
- events, webhooks, diagnostics, and observability.

Its infrastructure includes:

- **PostgreSQL** for durable product and control-plane state;
- **Redis** for admission, coordination, ephemeral distributed state, and rate
  limiting;
- **NATS JetStream** for durable asynchronous events and background work.

NATS is not part of the live audio path.

## Tool execution

Models do not receive unrestricted database or infrastructure access.

Agent tool calls flow through Leamout's tool runtime, which is responsible for
authorization, schema validation, timeouts, cancellation, auditing, and
execution against approved customer systems.

```text
Model
  |
  v
Tool request
  |
  v
Leamout Tool Runtime
  |
  +-- authorize
  +-- validate
  +-- execute
  +-- audit
  |
  v
Customer API / CRM / service
```

## Commercial boundary

Billing and payment processing are intentionally outside the current runtime
architecture while the commercial model is redesigned separately.

Customer carrier spend remains outside Leamout.

## Services

- `server`: HTTP control plane and public API.
- `worker`: asynchronous jobs, event consumers, and background coordination.
- `media`: low-latency realtime audio and AI-provider orchestration.
- `opensips`: SIP routing and policy enforcement for customer-provided trunks.
- `freeswitch`: call application and media runtime.
- `rtpengine`: RTP/media boundary.
- `coturn`: WebRTC STUN/TURN when required.

The Go codebase remains modular. Product modules should not become independent
network services unless scaling, isolation, or deployment requirements justify
that boundary.

## Repository direction

Current engineering work is focused on turning the existing telephony platform
into a unified autonomous voice-agent runtime:

```text
telephony runtime
      |
      v
realtime media
      |
      v
agent sessions
      |
      +-- turn control
      +-- barge-in
      +-- model orchestration
      +-- tool execution
      +-- human handoff
      +-- observability
```

General-purpose messaging, managed carrier commerce, telecom resale, telecom
wallet charging, and payment adapters are outside the current product boundary.

See [the Voice Agent execution plan](docs/voice-agent-execution-plan.md) for the
ordered delivery plan covering organization AI integrations, agent presets and
readiness, the golden call lifecycle, operational release gates, and the
product-facing API and UI.

## Validation

Run the server checks:

```sh
cd server

go test ./...
go vet ./...
go build ./...
```

Configure and validate the deployment from the repository root:

```sh
cp .env.example .env
cp server/.env.example server/.env
# Replace example secrets and addresses in both files before deploying.
docker compose --env-file .env -f deploy/compose.yaml config --quiet
docker compose --env-file .env -f deploy/compose.yaml up -d
```

The root `.env` supplies infrastructure settings for Compose, such as
`POSTGRES_PASSWORD` and `PUBLIC_IP`. `server/.env` supplies application settings
through `env_file`; Go owns their defaults and validation. The server, worker,
media runtime, Caddy, FreeSWITCH, and Coturn load the same application file so
shared domains and secrets stay consistent.

Keep the password in `server/.env`'s `DATABASE_URL` consistent with the root
`POSTGRES_PASSWORD`. Set `DOMAIN`, `CORS_ORIGINS`, `TURN_PUBLIC_URLS`, and the root
`TURN_REALM` for your domain. Container hostnames in the application example are
for Compose; replace them with reachable addresses when running Go directly.
Acceptance suites generate their own temporary application environment files.
To validate examples without creating deployment files, run
`APP_ENV_FILE=../server/.env.example docker compose --env-file .env.example -f deploy/compose.yaml config --quiet`.
Keep both environment files private. No `.env` file is copied into the image.

## Cloud billing with Polar

Cloud checkout and customer portal sessions use Polar. Configure `POLAR_ACCESS_TOKEN`
and `POLAR_WEBHOOK_SECRET` in the application environment. `POLAR_ENVIRONMENT`
defaults to `production`; use `sandbox` with sandbox credentials and product IDs
for acceptance testing. The API contract is pinned to `2026-10`.

The production product mappings default to:

| Plan | Polar product ID |
| --- | --- |
| Developer | `58ff657a-74d9-4342-89c4-2a3b784c6965` |
| Pro | `da5b3818-e92e-4b09-afad-ae61895f05db` |
| Scale | `a3023020-9822-4835-9509-1635c7d35f63` |

Override these with `POLAR_DEVELOPER_PRODUCT_ID`, `POLAR_PRO_PRODUCT_ID`, and
`POLAR_SCALE_PRODUCT_ID` when using another Polar environment. Configure recurring
products and portal upgrade/downgrade settings in Polar; local plan quotas remain
in Leamout. Checkout binds Polar's external customer ID to the Leamout organization
ID and maps the purchased product ID to a plan, independently of checkout metadata.

Configure a Polar webhook endpoint at `/v1/webhooks/polar` with API version
`2026-10` and subscription lifecycle events: created, updated, active, canceled,
uncanceled, cycled, revoked, past_due, paused, resumed, and migrated. The raw body
is verified against the webhook identity, timestamp and signature headers. Enable
`customer.state_changed` as well for immediate Customer State recovery. Both
new Standard Webhooks secrets and the documented legacy Polar signing scheme are
supported. Verified events are deduplicated by `webhook-id` in PostgreSQL before
Leamout responds, then a worker processes them with leases, exponential retries,
and dead-lettering. Repeated subscription notifications upsert the same local row,
and older observations cannot overwrite newer ones. Scheduled cancellation
preserves access until the current paid period ends. Customer State reconciliation
also checks organizations without a local subscription, allowing a missed initial
purchase event to be recovered.

When Polar reports `past_due`, Leamout keeps the current paid limits for seven days
by default. Set `POLAR_PAYMENT_GRACE_DAYS` to a value from 0 through 21 to change
this policy. The deadline is persisted when the delinquency begins, so webhook
retries and restarts cannot extend it. New calls use Free limits after the deadline,
while calls that are already active can finish.

The existing authenticated `/v1/billing/checkout` and `/v1/billing/portal` APIs
return Polar URLs. `/v1/billing/status` provides the console's plan, usage,
renewal, cancellation, and payment-grace state. After checkout, the console uses
`/v1/billing/reconcile` for bounded Customer State recovery while the webhook
worker catches up. Checkout still requires `Idempotency-Key` and uses Leamout's
local request-replay middleware. `cancel_url` is passed as Polar's `return_url`
for the checkout back button; success and return URLs must use the configured
frontend origin. No undocumented provider idempotency header is assumed.

Migration 034 creates the subscription schema with Polar as the supported billing
provider. Self-hosted installations continue using their licenses without a Polar
subscription or Cloud quotas.

## License

The software in this repository is licensed under the [Apache License 2.0](LICENSE).
See [NOTICE](NOTICE) for attribution information.

The Apache License 2.0 does not grant permission to use Leamout trade names,
trademarks, service marks, or product names except as permitted by the license.
Separately distributed Leamout Cloud, enterprise, or proprietary software may
be provided under different terms.
