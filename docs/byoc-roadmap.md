# BYOC roadmap

This roadmap tracks customer-owned SIP carrier connectivity independently from programmable-media completion.

## Phase 1 — secure carrier onboarding

- [x] Organization-scoped carrier connections and source-IP records.
- [x] Trunks with multiple SIP endpoints.
- [x] Encrypted outbound digest credential API.
- [x] Encrypted inbound digest or source-IP authentication API.
- [x] Public number-to-carrier-connection assignment API.
- [x] Add a built-in generic SIP provider to production bootstrap data.
- [x] Apply source-IP changes immediately (carrier ingress queries PostgreSQL per request).
- [x] Materialize encrypted digest credentials into an OpenSIPS-safe HA1 runtime store without restart.

### Credential activation design

Source-IP authentication is already live immediately because OpenSIPS resolves
the active CIDR records from PostgreSQL for every new carrier INVITE. Digest
credentials are different: OpenSIPS cannot consume the API's AES-GCM ciphertext
directly, and a generic `reload` command would not make those secrets usable.

Implement digest activation as a separate security change:

1. [x] Add an explicit SIP realm to inbound and outbound digest configuration.
2. [x] Derive and store realm-bound HA1 material; never
       expose or copy plaintext credentials into an OpenSIPS-readable table.
3. [x] Authenticate inbound carrier requests against the active, organization-bound
       HA1 records and resolve the same carrier connection after authentication.
4. [x] Add an outbound credential agent that uses the realm-bound HA1 material
       to answer carrier challenges without exposing plaintext.
5. [x] Rotate runtime material atomically with encrypted control-plane state.
6. [x] Add acceptance coverage proving set, rotate, and delete materialization without
       restarting OpenSIPS or exposing plaintext in SQL, logs, or MI responses.

Do not implement this item as an OpenSIPS configuration reload: the current SIP
route does not reference carrier digest credentials, so reload alone would
report success while leaving authentication behavior unchanged.

## Phase 2 — dedicated BYOC acceptance gate

- [x] Add an independent BYOC v1 Compose gate and CI workflow.
- [x] Provision a generic carrier exclusively through public APIs.
- [x] Complete inbound and outbound calls over UDP.
- [x] Reject cross-organization DID ownership (unknown source rejection is covered).
- [x] Exercise outbound digest authentication and credential rotation.
- [x] Persist and verify carrier connection and endpoint IDs in addition to trunk attribution.
- [x] Verify configuration survives API and OpenSIPS restarts.

## Phase 3 — endpoint reliability

- [x] Add SIP OPTIONS health checks and priority failover.
- [x] Distribute traffic by weight among equal-priority eligible endpoints.
- [x] Exclude unhealthy endpoints from weighted selection.
- [x] Add failure cooldowns and circuit-breaker state.
- [x] Expose endpoint health diagnostics and selected endpoint attribution.

## Phase 4 — secure transports and identity

- [x] Support SIP TLS with SNI, verified certificate chains, and hostname-verified health probes.
- [x] Support deployment-managed customer CA bundles and optional mutual TLS identity.
- [x] Enforce caller identity ownership and per-call privacy, DTMF, and codec-order policies.
- [x] Apply SDES-SRTP media policy independently from SIP TLS transport.

## Phase C — secure WebRTC

- [x] Issue tenant-bound, short-lived TURN credentials through an authenticated API.
- [x] Enable SIP over secure WebSocket at the OpenSIPS edge.
- [x] Add RTPengine WebRTC-to-SIP DTLS-SRTP and ICE policies.
- [x] Restore browser acceptance coverage with forced TURN relay.

The WebRTC v1 acceptance gate provisions its own SIP domain and subscriber through
public APIs, obtains short-lived ICE credentials from `/v1/realtime/ice-credentials`,
registers Chromium to OpenSIPS over WSS, and places a real audio call through
RTPengine and FreeSWITCH. Chromium is configured with `iceTransportPolicy: "relay"`.
The gate verifies that a TURN relay candidate is gathered, that the nominated media
path remains inside the Coturn relay-port range, and that inbound RTP audio bytes are
received. This prevents host or server-reflexive fallback from satisfying the test.

## Phase 5 — operations and orchestration

- [x] Enforce CPS, concurrent-call, and daily-minute limits.
  - [x] Coordinate CPS and concurrent-call admission atomically through Redis.
  - [x] Calculate daily answered seconds from durable call timestamps.
  - [x] Fail closed when Redis or durable usage state is unavailable.
  - [x] Release, renew, and reconstruct leases through lifecycle reconciliation.
- [x] Publish metrics per carrier connection, trunk, and endpoint.
  - [x] Count attempted, answered, failed, and completed calls.
  - [x] Count quota rejections and primary/failover endpoint selections.
  - [x] Publish endpoint probe results, health, and latency.
  - [x] Bound resource-attributed Prometheus series cardinality.
- [x] Add audit events for credential rotation and number reassignment.
  - [x] Record credential set, rotation, and deletion with redacted metadata.
  - [x] Record number assignment and reassignment with previous/new carrier IDs.
  - [x] Attribute append-only events to the actor, organization, target, and timestamp.
- [x] Add carrier test-call and configuration-validation workflows.
  - [x] Add an on-demand configuration validator with bounded SIP OPTIONS probes.
  - [x] Add a controlled, destination-allowlisted test-call workflow.
- [ ] Add multi-carrier policy only after single-carrier BYOC is reliable.

## Phase D — production hardening

- [x] Apply a shared Redis rate limit to tenant TURN credential issuance.
- [x] Include Redis in API readiness and fail closed when quota state is unavailable.
- [x] Mark credential responses as non-cacheable.
- [x] Add graceful drain controls for OpenSIPS, RTPengine, and FreeSWITCH nodes.
- [ ] Separate public signaling, public media, and private control networks.
- [ ] Add certificate and shared-secret rotation runbooks and automation.

## MVP exit criteria

BYOC v1 is complete when an organization can configure a generic SIP carrier, authenticate both directions, assign a DID, complete inbound and outbound calls, rotate credentials, and recover from component restarts without database intervention.
