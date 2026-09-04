# CommPeak provider

CommPeak is the initial managed provider for SIP termination and origination.

This document records how CommPeak should plug into the existing Leamout carrier model. It is an implementation note, not a provider overview.

## Phase 1 scope

Phase 1 needs to support:

- outbound PSTN termination through a CommPeak trunk;
- inbound CommPeak SIP traffic where origination is enabled;
- SIP authentication and source-IP attribution;
- endpoint health/failover through the existing trunk model;
- call attribution to carrier connection, trunk, endpoint, and SIP call ID;
- provider CDR ingestion for wholesale reconciliation.

Not in the first integration:

- multi-carrier LCR;
- wholesale-rate-driven route selection;
- provider-specific customer pricing.

## Existing Leamout model

Use the existing SIP carrier primitives.

| Leamout object | CommPeak use |
| --- | --- |
| `carrier_providers` | Built-in CommPeak provider definition. |
| `carrier_connections` | Organization-scoped SIP auth, limits, codecs, and ingress mode. |
| `carrier_connection_source_ips` | CommPeak source networks accepted for inbound traffic. |
| `trunks` | Logical CommPeak trunk. |
| `trunk_endpoints` | CommPeak SIP proxy targets. |
| `calls` | Persists carrier/trunk/endpoint attribution and `sip_call_id`. |

CommPeak is standards-based SIP. Use the existing SIP adapter convention unless provider-specific control-plane behavior later requires a separate adapter.

## Outbound path

Current outbound routing is trunk-scoped. Do not describe it as LCR.

Expected path:

```text
calls.Service
    ↓
routing.ResolveOutbound
    ↓
trunk
    ↓
carrier_connection
    ↓
trunk_endpoints
    ↓
priority / health / weight selection
    ↓
FreeSWITCH / OpenSIPS
    ↓ SIP INVITE
CommPeak
    ↓
PSTN
```

`calls.Service` already enforces carrier connection admission before origination. Keep CommPeak inside that path; do not add a provider-specific bypass around routing or admission control.

The selected route must persist:

```text
carrier_connection_id
trunk_id
trunk_endpoint_id
sip_call_id
```

Those fields are required later for debugging and CDR reconciliation.

## SIP authentication

Use the existing `carrier_connections` authentication fields.

For digest auth:

```text
outbound_auth_method = digest
auth_username
auth_secret_ciphertext
```

Leamout already materializes realm-bound HA1 runtime state for OpenSIPS. CommPeak credential rotation must use that path rather than introducing provider-specific plaintext configuration.

If the CommPeak account uses source-IP authentication instead, configure that explicitly. Do not assume password or IP auth globally for the provider.

## Endpoints

Represent every CommPeak target as a `trunk_endpoints` row.

Use the provider-assigned values for:

```text
host
port
transport
direction
priority
weight
```

Do not hard-code a sample CommPeak proxy hostname into production bootstrap data unless it is a provider-documented stable endpoint for all accounts.

The existing endpoint health logic is responsible for OPTIONS probes, cooldowns, priority failover, and weighted selection. CommPeak does not need its own health subsystem.

## Inbound path

Where CommPeak origination is used:

```text
CommPeak
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

Store authorized CommPeak ingress CIDRs in `carrier_connection_source_ips`.

Do not maintain a separate static CommPeak allow list in `opensips.cfg`. Source-IP changes should remain live through the existing database-backed ingress lookup.

Inbound routing must fail closed when:

- the source IP is unknown;
- the source matches more than one carrier connection ambiguously;
- the called number is assigned to another carrier connection;
- the organization does not match;
- no voice binding exists.

## Caller identity

Outbound caller identity must stay inside the existing ownership checks.

A number used as `From` must be voice-enabled and assigned to the same carrier connection as the selected trunk. Do not special-case CommPeak to allow arbitrary caller IDs through the API.

Provider-specific CLI restrictions can be added later as validation on top of the existing ownership rule.

## Codecs

Configure the connection with codecs supported by the active CommPeak account and by Leamout media services.

Prefer avoiding transcoding where possible. Codec policy belongs on the carrier connection / call policy path, not in a CommPeak-specific FreeSWITCH profile.

## Wholesale rates

`carrier_rates` is customer-facing billing data. Do not store CommPeak wholesale termination rates there.

Before cost-aware routing or margin accounting is implemented, add a separate upstream cost model.

It needs to handle at least:

```text
provider / carrier connection
destination prefix
country / network
route or pricing group
currency
unit cost
minimum / increment
effective period
rate deck version
```

Only after that model has deterministic precedence and effective dating should routing consider cost across multiple carriers.

## CDR ingestion

Provider CDRs are asynchronous reconciliation data. They must not be part of the synchronous call-control path.

Planned provider package:

```text
server/internal/integrations/carriers/commpeak/
```

The CommPeak client should expose only the operations needed to fetch termination/origination records and any rate data Leamout actually consumes.

A worker should:

1. fetch CDRs incrementally;
2. use a durable cursor or overlapping time window;
3. normalize provider records;
4. match them to internal calls;
5. persist wholesale duration/cost separately from customer billing;
6. record unmatched or conflicting records;
7. be safe to rerun without double-counting.

Use `calls.sip_call_id` for reconciliation only if the provider exposes a compatible call identifier. Verify that behavior against the production account before depending on it.

Do not assume a custom SIP header will appear in CommPeak CDRs unless this has been tested end to end.

## CDR storage

The current `calls` table does not own upstream billed cost fields.

Do not start adding provider-specific columns directly to `calls` for every carrier. Define a provider-neutral reconciliation model when implementation starts.

It will likely need:

```text
provider
provider CDR ID
call_id
billed seconds
wholesale amount
currency
provider status
raw/normalized timestamps
reconciled_at
```

The provider CDR ID must be unique enough to make ingestion idempotent.

## Implementation checklist

- [ ] Add the built-in `commpeak` carrier provider.
- [ ] Obtain the production SIP authentication mode and endpoint list.
- [ ] Create an organization-scoped CommPeak `carrier_connection`.
- [ ] Configure outbound auth through the existing encrypted credential path.
- [ ] Create the required trunk and `trunk_endpoints` rows.
- [ ] Verify OPTIONS health checks against CommPeak targets.
- [ ] Complete a real outbound call through the normal `calls.Service` path.
- [ ] Verify carrier connection, trunk, endpoint, and SIP call ID attribution.
- [ ] Rotate outbound credentials without restarting OpenSIPS.
- [ ] Populate `carrier_connection_source_ips` when inbound origination is enabled.
- [ ] Complete a real inbound CommPeak call through the existing number/binding path.
- [ ] Implement the CommPeak CDR client.
- [ ] Define provider-neutral wholesale reconciliation storage.
- [ ] Reconcile provider CDRs idempotently.
- [ ] Add acceptance coverage for outbound termination, auth rotation, endpoint failover, and inbound origination.

## Phase 1 exit criteria

CommPeak Phase 1 is complete when a test can:

1. provision the CommPeak carrier connection and endpoints;
2. originate a real PSTN call through CommPeak;
3. enforce CPS/concurrency/daily-minute admission on that call;
4. persist the selected connection, trunk, endpoint, and SIP call ID;
5. rotate credentials without an OpenSIPS restart;
6. fail over from an unhealthy primary endpoint;
7. accept and attribute a real inbound CommPeak INVITE when origination is enabled;
8. fetch the provider CDR for the call;
9. reconcile it exactly once to the internal call record / reconciliation record.
