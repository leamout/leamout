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

## Signing self-hosted licenses and releases

License signing and image signing use separate trust systems. Generate an Ed25519
license key on a protected issuance machine; keep the private key outside GitHub
Actions, source control, and customer images:

```sh
cd server
go run ./cmd/license -generate-key -output /secure/issuer-2026.private.key
```

The command prints the public key only and creates the private file with mode 0600.
Keep an encrypted backup with restricted access. Use `-claims`, `-private-key`,
`-key-id`, and `-output` to issue customer license files locally. The key ID must
contain only letters, digits, underscores, or hyphens, start with a letter or digit,
and be at most 64 characters. Never configure a private key as a release input.

Create the GitHub Actions environment `self-hosted-release`, restrict deployments
to release tags, and configure required reviewers. Set its **variable**
`LICENSE_PUBLIC_KEYS` to `issuer-2026:BASE64_PUBLIC_KEY`. Multiple public keys use
comma-separated entries. You can check this value before configuring it:

```sh
go run ./cmd/license -check-keyring -public-keys 'issuer-2026:BASE64_PUBLIC_KEY'
```

After merging the signing workflow, publish a GitHub release with a version tag
such as `v1.0.0` pointing to a commit on main. The workflow rejects missing,
malformed, duplicate, or private-key inputs before building. It publishes amd64
images at `ghcr.io/leamout/leamout-self-hosted/runtime` and
`ghcr.io/leamout/leamout-self-hosted/migrate`. The runtime image contains the API,
worker, and media binaries with the issuer public keys pinned at build time.
The migration image contains schema tooling and does not perform license checks.

Cosign uses the release workflow's GitHub OIDC identity to sign each published
image digest and immediately verify it. No long-lived image-signing key is stored
in GitHub. Build provenance and SBOM attestations are emitted by Buildx. The job
summary and artifacts list verified image digests; a failed signing step must not
be treated as a successful release, even if its image tag was already uploaded.
Do not rerun a version tag with different source or public keys: use a new version.

Customers should verify the exact digest before deployment:

```sh
cosign verify 'ghcr.io/leamout/leamout-self-hosted/runtime@sha256:DIGEST' \
  --certificate-identity 'https://github.com/leamout/leamout/.github/workflows/release-self-hosted.yml@refs/tags/v1.0.0' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com'
```

Use the same verification for the migration image. Deploy the verified digest,
not a mutable tag. Registry visibility/access must be configured separately for
customer distribution. Image signature verification is an installation step;
the runtime does not automatically verify its own container signature.

For rotation, first release binaries trusting both old and new public keys, then
issue licenses with the new private key. Retain the old public key until customers
have renewed and upgraded. If a license key is compromised, stop issuance with it
and ship a replacement keyring; offline installations cannot be remotely revoked.

## License

The software in this repository is licensed under the [Apache License 2.0](LICENSE).
See [NOTICE](NOTICE) for attribution information.

The Apache License 2.0 does not grant permission to use Leamout trade names,
trademarks, service marks, or product names except as permitted by the license.
Separately distributed Leamout Cloud, enterprise, or proprietary software may
be provided under different terms.
