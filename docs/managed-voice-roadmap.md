# Managed voice roadmap

This roadmap defines the transition from self-hosted programmable voice with BYOC to Leamout-managed programmable voice.

Managed voice must build on the existing carrier, trunk, routing, call, and number primitives. It must not make DIDWW or CommPeak architectural dependencies.

## Invariants

- Customer-facing calls, numbers, applications, events, and webhooks remain organization-scoped.
- BYOC remains a first-class routing mode after managed voice is added.
- A Leamout-managed carrier account may serve many organizations.
- Provider credentials and provider resource IDs remain internal control-plane state.
- Number provisioning and outbound termination are independent. A DID supplied by DIDWW may be used as caller identity on a call terminated through CommPeak.
- Provider adapters translate provider protocols; they do not own Leamout persistence, routing, tenancy, billing, or webhook delivery.
- Managed voice starts with one deterministic upstream route. Multi-carrier cost/quality routing comes later.

## Phase 1 — ownership foundation

The current carrier model is organization-owned:

```text
organization
    ↓
carrier_connection
    ↓
trunk
    ↓
trunk_endpoint
```

Managed voice also needs platform-owned connectivity:

```text
Leamout platform
    ↓
carrier_connection
    ↓
trunk
    ↓
trunk_endpoint
    ↓
many organizations
```

Do not create a second SIP stack with separate managed connection/trunk/endpoint tables. Generalize the existing carrier primitives so the connection owns the scope and the same auth, health, failover, metrics, and OpenSIPS runtime paths work for both modes.

Planned carrier-connection scopes:

```text
organization  # customer-owned/BYOC
platform      # Leamout-managed upstream
```

Rules:

- organization-scoped connections require an `organization_id`;
- platform-scoped connections have no tenant owner;
- public carrier-connection APIs only create/read organization-scoped connections;
- platform-scoped connections are operator/internal resources and are never exposed as another organization's BYOC connection;
- trunks, endpoints, source IPs, credentials, health state, and quotas remain attached to the carrier connection and inherit its scope.

The schema migration should normalize tenant ownership around `carrier_connections`. Existing redundant organization IDs on trunks/endpoints/source-IP rows must not become a second source of truth for platform scope.

## Phase 2 — number ownership

`phone_numbers` remains organization-scoped regardless of how the number was obtained.

A number needs an explicit provisioning mode:

```text
byoc
managed
```

A managed number additionally needs provider attribution independent from the SIP carrier used for outbound calls:

```text
provider_id
provider_resource_id
```

Rules:

- BYOC numbers are registered/assigned by the organization and use an organization-scoped ingress connection.
- Managed numbers are provisioned by Leamout through a provider adapter and may use a platform-scoped ingress connection.
- The organization owns the Leamout `phone_numbers` resource; Leamout owns the upstream provider account/resource.
- `(provider_id, provider_resource_id)` identifies the upstream resource when present.
- Provider-specific capacity, routing-resource IDs, status, and wholesale cost must not be packed into unrelated phone-number columns.

## Phase 3 — routing behavior

### Outbound

Today an outbound call requires an organization-owned `trunk_id`.

Managed voice needs two route sources:

```text
BYOC
  explicit organization trunk

managed
  Leamout-selected platform trunk
```

The public call API should evolve so `trunk_id` is optional:

- `trunk_id` present: resolve only an organization-scoped BYOC trunk owned by the caller's organization;
- `trunk_id` absent: resolve the current managed route policy;
- do not silently fall from a requested BYOC trunk to managed voice;
- do not silently fall from managed voice to a customer's BYOC trunk.

The first managed route policy may select one configured platform trunk. It is not LCR.

### Caller identity

Current BYOC behavior requires the caller-ID number and outbound trunk to share the same carrier connection. Keep that rule for BYOC.

Managed voice must not require the number provider and termination provider to match.

For a managed outbound route, caller identity is authorized when the number:

- belongs to the calling organization;
- is active and voice-enabled;
- is allowed for outbound caller identity by Leamout policy.

This permits:

```text
DIDWW managed DID
    ↓ caller identity
Leamout managed route
    ↓
CommPeak
```

### Inbound

Inbound routing must derive tenant ownership from the called number when the ingress carrier connection is platform-scoped.

```text
provider source IP
    ↓
carrier connection
    ↓
called number
    ↓
phone_numbers.organization_id
    ↓
voice binding
```

Rules:

- the called number must be assigned to the resolved ingress carrier connection;
- organization-scoped connections additionally require the number organization to match the connection organization;
- platform-scoped connections do not supply the tenant identity;
- the voice binding must belong to the phone number's organization;
- the routing decision organization is the phone number's organization.

## Phase 4 — managed provider integration

Only after the ownership/routing foundation exists should provider adapters enter customer workflows.

### DIDWW

1. provider HTTP client/auth;
2. number search;
3. number order;
4. persist provider attribution;
5. configure DID routing to a platform-scoped Leamout ingress connection;
6. reconcile provider inventory/routing state.

### CommPeak

1. configure a platform-scoped SIP connection/trunk through the generic carrier path;
2. use it as the first managed outbound route;
3. ingest termination/origination CDRs asynchronously;
4. add provider-neutral wholesale reconciliation storage;
5. ingest wholesale rates only after the upstream-cost model exists.

## Phase 5 — coexistence gate

Before multi-carrier routing, acceptance coverage must prove all of these simultaneously:

- organization A places and receives calls through BYOC only;
- organization B buys a managed DID and receives calls through managed ingress;
- organization B originates a managed outbound call without supplying a trunk;
- a managed DID can be caller identity on a different managed termination carrier;
- a requested BYOC trunk never falls back to managed voice;
- managed routing never consumes another organization's BYOC connection;
- provider credentials/resource IDs are not exposed through customer APIs;
- restart/reconciliation preserves route ownership and call attribution.

## Later — multi-carrier managed voice

Do not add this until the coexistence gate is reliable.

Future work includes:

```text
multiple managed termination carriers
multiple numbering providers
provider-neutral wholesale rate decks
quality metrics
routing policies
LCR
carrier failover across providers
margin accounting
direct interconnects
```

DIDWW and CommPeak remain initial adapters, not permanent routing assumptions.
