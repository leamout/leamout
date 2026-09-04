# CommPeak Integration

CommPeak is Leamout's primary wholesale provider for **SIP voice termination (outbound)** and **origination (inbound)**.

CommPeak exposes standards-based SIP connectivity for call routing, while provider-side rate and CDR data can be used by Leamout for reconciliation and wholesale cost tracking. CommPeak currently advertises SIP trunking from approximately `$0.0009/min`, but actual rates vary by route, pricing group, destination, billing increment, account tier, and negotiated terms and must not be hard-coded into Leamout.

## 1. Strategic Role in Leamout

- **Primary use case**: Terminate outbound calls from Leamout customers to the global PSTN.
- **Secondary use case**: Receive inbound PSTN traffic through CommPeak origination where CommPeak-managed DIDs or inbound routes are used.
- **Why CommPeak**: Wholesale-oriented SIP termination, broad international routing, standards-based SIP interoperability, and provider CDR/rate data suitable for cost reconciliation.

## 2. Integration Architecture

CommPeak call signaling integrates with Leamout through SIP. Operational and billing data such as rates and CDRs may be retrieved separately from CommPeak's portal or APIs, but Leamout does not depend on a proprietary REST call-control API to originate or receive calls.

### A. Outbound Termination

1. `calls.Service` receives an outbound call request.
2. The telecom routing layer resolves the requested trunk, its `carrier_connection`, and an eligible outbound `trunk_endpoint`.
3. Admission control enforces the carrier connection's CPS, concurrent-call, and optional daily-minute limits.
4. The media/signaling controller originates the call toward the selected CommPeak SIP endpoint.
5. Outbound SIP authentication is applied according to the CommPeak account configuration, typically SIP Digest or provider-authorized source IP.
6. CommPeak terminates the call to the PSTN.

The current routing resolver performs endpoint selection within the requested trunk using endpoint priority, health, and weight. Least Cost Routing (LCR) across multiple carriers is a future routing capability and should not be documented as present until a provider-cost routing model exists.

### B. Inbound Origination

1. A PSTN call reaches a CommPeak DID or origination route configured for Leamout.
2. CommPeak sends a SIP INVITE to Leamout's public SIP ingress.
3. OpenSIPS resolves the source IP against `carrier_connection_source_ips` and attributes the INVITE to the corresponding organization-scoped `carrier_connection`.
4. The routing layer verifies that the called `phone_numbers` record belongs to the same carrier connection.
5. The number is resolved through `voice_bindings` to the target voice application.

Inbound source authentication should use the actual CommPeak source networks assigned to the account rather than a static global allow list embedded in OpenSIPS configuration.

## 3. Leamout Data Model Mapping

| Leamout entity | CommPeak concept | Notes |
| :--- | :--- | :--- |
| `carrier_providers` | CommPeak platform | Add a provider such as `slug: "commpeak"`. The current standards-based carrier adapter convention is `adapter: "sip"`. |
| `carrier_connections` | SIP account / carrier relationship | Stores organization-scoped SIP auth configuration, inbound mode, traffic limits, codecs, and provider relationship. For password auth use `outbound_auth_method: "digest"`, `auth_username`, and encrypted `auth_secret_ciphertext`. |
| `carrier_connection_source_ips` | Authorized CommPeak ingress networks | Stores CIDRs used to identify and authorize inbound SIP traffic. |
| `trunks` | CommPeak trunk | Represents the logical inbound, outbound, or bidirectional trunk attached to the CommPeak carrier connection. |
| `trunk_endpoints` | CommPeak SIP proxy / ingress target | Stores `host`, `port`, `transport`, direction, priority, weight, enabled state, and endpoint health information. |
| `calls` | Provider-routed call | Stores `carrier_connection_id`, `trunk_id`, `trunk_endpoint_id`, and `sip_call_id`, providing the internal attribution needed for later CDR reconciliation. |
| `carrier_rates` | Customer-facing billable rate | The current table is **not** the CommPeak wholesale rate table. It represents what Leamout charges customers and must not be overloaded with upstream carrier cost. |

### Wholesale cost model

Leamout does not currently have a dedicated upstream carrier-cost model. CommPeak termination rates therefore require a separate future model rather than reusing `carrier_rates`.

A provider-cost model should eventually support at least:

- carrier provider or carrier connection
- destination prefix
- country and network
- route or pricing group / technical prefix
- currency
- wholesale unit cost
- billing minimum and increment, such as `1/1`, `60/1`, or `60/60`
- effective start and end times
- source/version of the imported rate deck

This model can later feed LCR and margin analysis while remaining independent from customer pricing.

## 4. Billing and CDR Reconciliation

CommPeak provides call-detail reporting for termination and origination. Leamout should use provider CDR data to reconcile actual upstream billed duration and cost after the call completes.

### Preferred ingestion path

Use a background worker to pull CDRs from the provider's supported API or another account-approved export mechanism.

A possible integration location is:

```text
server/internal/integrations/carriers/commpeak/
```

with CDR reconciliation kept in a reusable carrier workflow rather than coupling it directly to call control.

The worker should:

1. Maintain a durable provider cursor or overlapping time window so delayed CDRs are not lost.
2. Pull termination and origination records incrementally.
3. Normalize provider fields such as source, destination, status, duration, billed duration, route/pricing group, rate, and cost.
4. Correlate the CDR to the internal `calls` record using a deterministic provider identifier or SIP identifier where the provider exposes one.
5. Make reconciliation idempotent so reprocessing the same provider CDR cannot double-count cost.
6. Persist provider-reported wholesale cost separately from customer-facing rated charges.
7. Record unmatched and conflicting CDRs for operational review instead of silently discarding them.

The existing `calls.sip_call_id` is unique and should participate in reconciliation when CommPeak exposes a compatible SIP call identifier. Do not assume a custom header such as `X-Leamout-Call-ID` will be returned in CommPeak CDRs until that behavior is verified with the production SIP account.

### Portal exports and SFTP

CommPeak supports CDR viewing and CSV export through its portal, and current CommPeak documentation exposes API endpoints for termination and origination call records. If a specific Leamout account is provisioned with SFTP delivery, it may be supported as an additional ingestion transport, but the integration should not depend on SFTP as a universal CommPeak capability.

## 5. Operational Requirements

1. **SIP authentication**: Obtain the CommPeak SIP account configuration. Store password credentials encrypted at rest when using digest authentication; use IP authentication only when explicitly configured on the provider account.
2. **Inbound source validation**: Populate `carrier_connection_source_ips` with the CommPeak ingress networks assigned or documented for the account.
3. **Endpoint configuration**: Represent each CommPeak SIP target as a `trunk_endpoints` row with the actual host, port, transport, direction, priority, and weight provided for the account.
4. **Codec negotiation**: Prefer codecs supported by both Leamout and the active CommPeak SIP service. CommPeak documents support for G.711 u-law/A-law and G.729a among other codecs. Avoid unnecessary transcoding when possible.
5. **Admission and fraud controls**: Enforce `max_cps`, `max_concurrent_calls`, and optional `max_daily_minutes` before an outbound call reaches the provider.
6. **Caller identity**: Ensure outbound caller IDs are authorized for the selected carrier connection and comply with CommPeak account and destination requirements.
7. **Rate synchronization**: Treat provider rate decks as versioned wholesale inputs. Do not assume the advertised starting price applies to every destination.
8. **CDR reconciliation**: Monitor ingestion lag, unmatched CDRs, duplicates, provider/API failures, and differences between expected and actual wholesale cost.
9. **Endpoint health**: Use Leamout's endpoint health state and priority/weight routing to fail over between configured CommPeak targets without routing to known unhealthy endpoints.

## 6. Implementation Checklist

- [ ] Create the built-in CommPeak entry in `carrier_providers` using the standards-based SIP adapter.
- [ ] Obtain the production CommPeak SIP account configuration and approved authentication method.
- [ ] Create the organization-scoped CommPeak `carrier_connection` with appropriate traffic limits and codecs.
- [ ] Create outbound/bidirectional `trunks` and CommPeak `trunk_endpoints` using provider-assigned hosts, ports, and transports.
- [ ] Verify outbound SIP Digest or IP authentication according to the CommPeak account configuration.
- [ ] Populate `carrier_connection_source_ips` for authorized CommPeak inbound source networks.
- [ ] Verify OpenSIPS attributes inbound CommPeak INVITEs to the correct organization and carrier connection.
- [ ] Add acceptance coverage for outbound CommPeak termination and endpoint failover.
- [ ] Add acceptance coverage for inbound CommPeak origination where a CommPeak DID/route is used.
- [ ] Implement a CommPeak CDR client for incremental termination/origination record retrieval.
- [ ] Define a dedicated upstream provider-cost schema instead of storing CommPeak wholesale costs in `carrier_rates`.
- [ ] Implement idempotent CDR-to-call reconciliation and unmatched-CDR handling.
- [ ] Add provider rate-deck import/versioning before enabling cost-aware LCR.
- [ ] Add LCR across carrier connections only after wholesale rates, precedence, effective dating, and failover policy are explicitly modeled and tested.

## 7. Phase 1 Definition of Done

The CommPeak integration is ready for Phase 1 when Leamout can:

1. Represent CommPeak as a built-in SIP carrier provider.
2. Configure a CommPeak carrier connection and one or more SIP endpoints.
3. Originate an outbound call through `calls.Service` and the existing routing/media controller path.
4. Authenticate successfully to CommPeak using the configured production method.
5. Attribute the persisted `calls` record to the selected carrier connection, trunk, endpoint, and SIP call ID.
6. Enforce carrier-level CPS, concurrency, and daily-minute controls before origination.
7. Accept and attribute inbound CommPeak traffic by source IP when origination is enabled.
8. Resolve inbound numbers to the correct voice application without cross-organization routing.
9. Retrieve provider CDRs and reconcile billed duration and wholesale cost idempotently.
10. Surface reconciliation failures without modifying customer-facing billing data incorrectly.

## 8. Provider References

- CommPeak pricing: https://www.commpeak.com/pricing
- CommPeak SIP origination and termination overview: https://docs.commpeak.com/docs/what-is-sip-origination-sip-termination
- CommPeak supported codecs: https://docs.commpeak.com/docs/what-are-the-supported-codecs
- CommPeak termination CDR API: https://docs.commpeak.com/reference/termination_call_records
- CommPeak origination CDR API: https://docs.commpeak.com/reference/origination_call_records
- CommPeak CDR portal/export documentation: https://docs.commpeak.com/docs/call-records-cdr
