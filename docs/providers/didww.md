# DIDWW Integration

DIDWW is Leamout's primary wholesale provider for **global phone number (DID) inventory**, **inbound SIP routing**, and **SMS capabilities**.

Unlike retail CPaaS providers, DIDWW operates on a wholesale model that is well suited to Leamout's carrier-agnostic architecture and cost structure.

## 1. Strategic Role in Leamout

- **Primary use case**: Source local, national, mobile, and toll-free numbers for Leamout customers.
- **Secondary use case**: Support inbound SMS termination where available.
- **Why DIDWW**: Broad global coverage, API-driven number management, wholesale pricing, and multiple capacity models for inbound voice.

## 2. Integration Architecture

DIDWW integrates with Leamout through two distinct planes.

### A. Control Plane (REST API)

Leamout uses the DIDWW v3 API to:

1. Search for available numbers by country, region, or area code.
2. Purchase numbers programmatically.
3. Configure the purchased DID to route inbound traffic toward Leamout-managed SIP infrastructure.
4. Reconcile provider-side inventory and state with Leamout's `phone_numbers` records.

Provider API credentials should be stored in Leamout's secrets infrastructure and referenced by the integration layer rather than exposed through SIP-facing carrier configuration.

### B. Data Plane (SIP Signaling)

- **Inbound calls**: DIDWW receives the PSTN call and forwards it to the SIP destination configured for the DID or associated inbound trunk.
- **Leamout ingress**: The SIP INVITE terminates at Leamout's OpenSIPS edge, where the source is attributed to the correct carrier connection and organization before normal call routing begins.
- **Media path**: RTP handling remains part of Leamout's media architecture through RTPengine and downstream media services. Provider-specific media behavior should not be assumed by the control-plane integration.

## 3. Leamout Data Model Mapping

| Leamout entity | DIDWW concept | Notes |
| :--- | :--- | :--- |
| `carrier_providers` | DIDWW platform | Add a provider entry such as `slug: "didww"` with an adapter identifier for the DIDWW integration. |
| `carrier_connections` | SIP connectivity configuration | Represents the organization-scoped DIDWW SIP relationship used for inbound attribution, limits, codecs, and carrier routing. REST API credentials should remain in the integration/secrets layer. |
| `phone_numbers` | DID inventory | Store the DID in `number` and the DIDWW resource identifier in `provider_resource_id`. DIDWW-specific capacity and routing metadata require integration metadata or future schema support. |
| `voice_bindings` | Number assignment | Links a `phone_numbers.id` to the target `voice_application_id` for inbound application routing. |

### DIDWW-specific metadata

The current `phone_numbers` schema already provides `provider_resource_id`, `carrier_connection_id`, and voice/SMS capability flags, but it does not currently expose dedicated columns for DIDWW capacity type or provider routing URI.

The DIDWW integration should therefore treat the following as provider-specific metadata until the schema is extended:

- DIDWW resource ID
- capacity model
- provider routing configuration
- provider-side status
- recurring and usage cost inputs required by commercial rating

If these fields become first-class operational inputs, add explicit schema support rather than overloading unrelated columns.

## 4. Capacity and Pricing Models

DIDWW supports different inbound voice capacity models. Leamout should persist the selected model for each DID or associated provider resource so wholesale cost can be represented accurately in the commercial rating pipeline.

Typical models include:

- **DID+0**: The DID itself does not include bundled voice channels. Capacity is purchased or billed separately. This is useful when traffic is variable or when capacity is managed independently from individual numbers.
- **DID+2**: The DID includes a small amount of bundled channel capacity, with additional capacity handled separately. This can fit predictable low-volume deployments.

Pricing, minimum commitments, tier thresholds, channel charges, and per-minute rates are commercial inputs and can change over time. They should be sourced from the active DIDWW account or current DIDWW pricing data rather than hard-coded into the integration.

The commercial layer should be able to distinguish at least:

- number monthly recurring cost (MRC)
- number non-recurring cost (NRC), when applicable
- capacity model and capacity charges
- inbound usage charges
- SMS charges
- provider tier or negotiated pricing context

## 5. Operational Requirements

1. **SIP source validation**: Maintain DIDWW source networks in the relevant `carrier_connection_source_ips` records so OpenSIPS can attribute and authorize inbound carrier traffic.
2. **Traffic controls**: Configure appropriate `max_cps`, `max_concurrent_calls`, and optional daily-minute limits on the DIDWW carrier connection.
3. **Secrets management**: Keep DIDWW REST credentials encrypted and outside SIP runtime configuration.
4. **Inventory reconciliation**: Periodically compare DIDWW-owned inventory with Leamout `phone_numbers` state and detect missing, released, or externally modified resources.
5. **Routing reconciliation**: Verify that provider-side SIP routing still points to the expected Leamout ingress configuration.
6. **Porting (LNP)**: Treat number porting as a separate workflow. Phase 1 can remain operational/manual until Leamout implements provider-assisted porting automation and document handling.
7. **SMS ingress**: Normalize DIDWW SMS callbacks or SMPP traffic into Leamout's messaging event model before exposing events to applications.

## 6. Implementation Checklist

- [ ] Create the built-in DIDWW entry in `carrier_providers`.
- [ ] Provision DIDWW API credentials and store them in Leamout's secrets manager.
- [ ] Implement `server/internal/integrations/carriers/didww/` client support for inventory search, ordering, resource lookup, and routing configuration.
- [ ] Define provider-specific metadata needed for DIDWW number capacity and wholesale cost tracking.
- [ ] Build a background reconciliation worker for DIDWW inventory and routing state.
- [ ] Create or update the organization-scoped DIDWW `carrier_connection` used for SIP ingress.
- [ ] Populate `carrier_connection_source_ips` with approved DIDWW source networks.
- [ ] Verify OpenSIPS accepts and attributes inbound DIDWW INVITEs correctly.
- [ ] Bind provisioned DIDs to voice applications through `voice_bindings`.
- [ ] Add acceptance coverage for inbound DIDWW call routing.
- [ ] Add SMS ingress support when messaging integration enters scope.

## 7. Phase 1 Definition of Done

The DIDWW integration is ready for Phase 1 when Leamout can:

1. Authenticate to DIDWW from the control plane.
2. Search available DID inventory.
3. Purchase a DID and persist the provider resource identifier.
4. Configure inbound SIP routing toward Leamout.
5. Attribute DIDWW SIP ingress to the correct organization and carrier connection.
6. Resolve the called DID to a Leamout `phone_numbers` record.
7. Route that DID through `voice_bindings` to the configured voice application.
8. Reconcile provider-side inventory without corrupting customer-owned control-plane state.
