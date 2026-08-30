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
- [ ] Materialize encrypted digest credentials into an OpenSIPS-safe runtime auth store without restart.

### Credential activation design

Source-IP authentication is already live immediately because OpenSIPS resolves
the active CIDR records from PostgreSQL for every new carrier INVITE. Digest
credentials are different: OpenSIPS cannot consume the API's AES-GCM ciphertext
directly, and a generic `reload` command would not make those secrets usable.

Implement digest activation as a separate security change:

1. Add an explicit SIP realm to inbound and outbound digest configuration.
2. Derive and store realm-bound HA1 material for inbound authentication; never
   expose or copy plaintext credentials into an OpenSIPS-readable table.
3. Authenticate inbound carrier requests against the active, organization-bound
   HA1 records and resolve the same carrier connection after authentication.
4. Add an outbound credential agent that decrypts credentials in trusted
   application memory and updates FreeSWITCH/OpenSIPS runtime authentication
   state atomically.
5. Rotate by installing new runtime material before retiring the previous
   version, then verify with an authenticated test transaction.
6. Add acceptance coverage proving set, rotate, and delete take effect without
   restarting OpenSIPS or exposing plaintext in SQL, logs, or MI responses.

Do not implement this item as an OpenSIPS configuration reload: the current SIP
route does not reference carrier digest credentials, so reload alone would
report success while leaving authentication behavior unchanged.

## Phase 2 — dedicated BYOC acceptance gate

- [ ] Provision a generic carrier exclusively through public APIs.
- [ ] Complete inbound and outbound calls over UDP.
- [ ] Reject unknown source IPs and cross-organization DID ownership.
- [ ] Exercise outbound digest authentication and credential rotation.
- [ ] Verify route attribution records carrier connection, trunk, and endpoint IDs.
- [ ] Verify configuration survives API, worker, and OpenSIPS restarts.

## Phase 3 — endpoint reliability

- [ ] Add SIP OPTIONS health checks and priority failover.
- [ ] Distribute traffic by weight among equal-priority healthy endpoints.
- [ ] Add failure cooldowns and circuit-breaker state.
- [ ] Expose endpoint health and route-attempt diagnostics.

## Phase 4 — secure transports and identity

- [ ] Support SIP TLS with hostname verification.
- [ ] Support customer CA bundles and optional mutual TLS.
- [ ] Add caller identity, privacy, DTMF, and codec-order policies.
- [ ] Add SRTP policy independently from SIP TLS.

## Phase 5 — operations and orchestration

- [ ] Enforce CPS, concurrent-call, and daily-minute limits.
- [ ] Publish metrics per carrier connection, trunk, and endpoint.
- [ ] Add audit events for credential rotation and number reassignment.
- [ ] Add carrier test-call and configuration-validation workflows.
- [ ] Add multi-carrier policy only after single-carrier BYOC is reliable.

## MVP exit criteria

BYOC v1 is complete when an organization can configure a generic SIP carrier, authenticate both directions, assign a DID, complete inbound and outbound calls, rotate credentials, and recover from component restarts without database intervention.
