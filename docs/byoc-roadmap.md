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

- [ ] Add SIP OPTIONS health checks and priority failover.
- [x] Distribute traffic by weight among equal-priority eligible endpoints.
- [ ] Exclude unhealthy endpoints from weighted selection.
- [ ] Add failure cooldowns and circuit-breaker state.
- [ ] Expose endpoint health and route-attempt diagnostics.

## Phase 4 — secure transports and identity

- [ ] Support SIP TLS with hostname verification.
- [ ] Support customer CA bundles and optional mutual TLS.
- [ ] Add caller identity, privacy, DTMF, and codec-order policies.
- [ ] Add SRTP policy independently from SIP TLS.

## Phase C — secure WebRTC

- [x] Issue tenant-bound, short-lived TURN credentials through an authenticated API.
- [ ] Enable SIP over secure WebSocket at the OpenSIPS edge.
- [ ] Add RTPengine WebRTC-to-SIP DTLS-SRTP and ICE policies.
- [ ] Add browser acceptance coverage with forced TURN relay.

## Phase 5 — operations and orchestration

- [ ] Enforce CPS, concurrent-call, and daily-minute limits.
- [ ] Publish metrics per carrier connection, trunk, and endpoint.
- [ ] Add audit events for credential rotation and number reassignment.
- [ ] Add carrier test-call and configuration-validation workflows.
- [ ] Add multi-carrier policy only after single-carrier BYOC is reliable.

## Phase D — production hardening

- [x] Apply a shared Redis rate limit to tenant TURN credential issuance.
- [x] Include Redis in API readiness and fail closed when quota state is unavailable.
- [x] Mark credential responses as non-cacheable.
- [ ] Add graceful drain controls for OpenSIPS, RTPengine, and FreeSWITCH nodes.
- [ ] Separate public signaling, public media, and private control networks.
- [ ] Add certificate and shared-secret rotation runbooks and automation.

## MVP exit criteria

BYOC v1 is complete when an organization can configure a generic SIP carrier, authenticate both directions, assign a DID, complete inbound and outbound calls, rotate credentials, and recover from component restarts without database intervention.
