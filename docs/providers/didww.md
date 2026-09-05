# DIDWW provider

DIDWW is the initial managed provider for DID inventory and inbound PSTN routing.

This document records the integration contract and the work required to make DIDWW a first-class Leamout provider. It is not a customer-facing provider overview.

## Phase 1 scope

Phase 1 needs to support:

- search DID inventory;
- order a DID;
- persist the DID and DIDWW resource ID;
- configure DIDWW to send inbound SIP traffic to Leamout;
- attribute DIDWW ingress to the correct organization and carrier connection;
- route the called number through `voice_bindings`;
- reconcile provider inventory and routing state.

Not in the first voice integration:

- automated porting;
- provider-specific customer billing logic;
- SMS delivery unless the messaging work requires it.

## Existing Leamout model

Use the existing carrier and number primitives.

| Leamout object | DIDWW use |
| --- | --- |
| `carrier_providers` | Built-in DIDWW provider definition. |
| `carrier_connections` | Organization-scoped DIDWW SIP ingress configuration. |
| `carrier_connection_source_ips` | DIDWW source networks accepted for the connection. |
| `phone_numbers` | Purchased DIDs. `provider_resource_id` stores the DIDWW resource ID. |
| `voice_bindings` | Maps a DID to a voice application. |

The current `phone_numbers` schema does not contain DIDWW capacity or provider-routing fields. Do not overload unrelated columns to store them.

Before the integration depends on those values at runtime, define storage for at least:

```text
capacity model
provider routing target / routing resource
provider-side status
wholesale MRC/NRC/usage inputs
```

## Provider definition

Add DIDWW to production bootstrap data when the adapter is usable.

The provider needs a stable slug:

```text
didww
```

The adapter value should identify the DIDWW control-plane integration. Do not reuse `sip` for REST inventory operations if adapter dispatch will use this value to select provider-specific code.

## Control-plane adapter

Planned package:

```text
server/internal/integrations/carriers/didww/
```

Keep DIDWW REST credentials out of `carrier_connections`. That table is SIP-facing carrier state and already owns SIP authentication, limits, codecs, and ingress configuration.

The DIDWW client needs the minimum operations required by number lifecycle:

```text
SearchNumbers
OrderNumber
GetNumber
ConfigureRouting
ReleaseNumber
```

Exact names are implementation details, but the adapter boundary should stay small and provider-specific.

### Provisioning flow

```text
API / worker
    ↓
DIDWW adapter
    ↓
search / order
    ↓
phone_numbers
    ↓
configure provider routing
    ↓
voice binding
```

A successful order is not complete until Leamout has persisted enough provider state to retry routing configuration safely.

Do not make number ordering and provider routing one irreversible operation. A failed routing update after a successful order must be recoverable by reconciliation.

## SIP ingress

DIDWW sends inbound calls to the Leamout SIP edge.

Expected path:

```text
DIDWW
  ↓ SIP INVITE
OpenSIPS
  ↓ source-IP carrier resolution
carrier_connection
  ↓ called number
phone_numbers
  ↓
voice_bindings
  ↓
voice application
```

Source-IP authorization is already data driven. Store the active DIDWW CIDRs in `carrier_connection_source_ips`; do not hard-code a permanent provider allow list into `opensips.cfg`.

Ingress must fail closed when:

- the source IP does not resolve to one carrier connection;
- the DID is not assigned to that carrier connection;
- the DID belongs to another organization;
- no voice binding exists for the called number.

## DIDWW routing state

Leamout needs one canonical SIP ingress target for managed DIDs per deployment or region.

The DIDWW adapter is responsible for translating that target into the provider's routing resources. Provider object IDs must remain provider metadata; OpenSIPS should not need to know them.

Routing reconciliation must detect at least:

- DID exists at DIDWW but not in Leamout;
- DID exists in Leamout but no longer exists at DIDWW;
- DIDWW routing no longer points at the expected Leamout ingress;
- provider status changed outside Leamout.

Do not silently delete local state during reconciliation. Report drift and make destructive recovery explicit.

## Capacity and wholesale cost

DIDWW capacity selection is provider state, not customer pricing.

The integration must preserve the selected capacity model (`DID+0`, `DID+2`, or any later provider variant) once Leamout starts using it for provisioning or cost calculations.

`usage_rates` is customer-facing commercial pricing. Do not store DIDWW wholesale cost there.

Wholesale number cost needs its own model when margin accounting is implemented. At minimum it will need to represent:

```text
provider
provider resource
currency
MRC
NRC
usage/capacity charges
effective period
```

## Reconciliation worker

The worker should be idempotent and safe to run repeatedly.

It must not treat provider state as authoritative for organization ownership or voice bindings. Those remain Leamout control-plane state.

Suggested responsibilities:

1. list or fetch Leamout-managed DIDWW resources;
2. compare provider status with `phone_numbers`;
3. verify the expected provider routing target;
4. retry incomplete routing configuration;
5. emit/record drift that requires operator action.

## Porting

Porting is manual in Phase 1.

Do not add an API that implies automated LNP until Leamout owns the full workflow for LOA/document collection, provider submission, status transitions, and failure handling.

## SMS

DIDWW SMS is outside the initial voice path.

When messaging work starts, normalize provider callbacks/SMPP traffic into the Leamout messaging model. Do not route raw DIDWW payloads directly to customer webhooks.

## Implementation checklist

- [ ] Add the built-in `didww` carrier provider.
- [ ] Add DIDWW REST credentials to the deployment secrets path.
- [ ] Implement the DIDWW control-plane adapter.
- [ ] Search available numbers.
- [ ] Order a number and persist `provider_resource_id`.
- [ ] Define storage for DIDWW-specific capacity and routing metadata required at runtime.
- [ ] Configure purchased DIDs to route to Leamout SIP ingress.
- [ ] Create the organization-scoped DIDWW carrier connection.
- [ ] Populate `carrier_connection_source_ips` with the active DIDWW ingress ranges.
- [ ] Route inbound DIDWW calls through the existing number/binding path.
- [ ] Add reconciliation for inventory and routing drift.
- [ ] Add acceptance coverage for purchase → route → inbound call.

## Phase 1 exit criteria

DIDWW Phase 1 is complete when a test can:

1. provision DIDWW credentials from deployment secrets;
2. search and order a DID;
3. persist the DID with its provider resource ID;
4. configure the DID to Leamout ingress;
5. receive a real DIDWW SIP INVITE;
6. attribute it to the expected carrier connection and organization;
7. resolve the DID to the expected voice application;
8. rerun reconciliation without duplicating or corrupting state.
