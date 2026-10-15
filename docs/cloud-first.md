# Leamout Cloud architecture

Leamout Cloud is the only supported Leamout product and deployment model.
Leamout operates one multi-tenant SIP-native infrastructure platform.

The application has three entrypoints: `server` (control-plane HTTP API),
`worker` (durable asynchronous tasks and orchestration), and `media`
(realtime call audio). OpenSIPS, FreeSWITCH, RTPengine, Coturn, PostgreSQL,
Redis, NATS JetStream and object storage provide the shared infrastructure.

## Commercial invariants

Every organization is subject to authenticated tenancy boundaries, BYOC
trunk admission, plans, runtime usage quotas and applicable feature
entitlements. Server-initiated and worker-initiated calls must enforce
the same plan policies. Runtime allowance exhaustion prevents **new**
sessions and does not terminate existing calls.

There is no deployment edition selector, signed-license runtime gate, or
unmetered customer Self-Hosted mode. A missing subscription may map to the
published Cloud Free tier, but missing/incomplete tier configuration must
fail closed rather than bypass limits.

## Deployments

`deploy/`, `containers/` and the acceptance test stacks are maintained
by Leamout to build, test and operate its hosted service. They are not
customer self-installation instructions.

## Source licensing

The current public repository contains Apache-2.0 licensed code. Existing
grants to distributed copies cannot be revoked merely by changing the
product's hosting strategy or making future development private. Changes to
future distribution and licensing require separate legal and repository
administration decisions.
