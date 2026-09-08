# DIDWW provider

DIDWW is the initial Leamout-managed provider for DID inventory, acquisition, and inbound PSTN routing.

DIDWW provider identifiers, credentials, Voice IN trunks, and other wholesale resources are internal Leamout infrastructure. They are not customer-facing API resources and they are not part of ordinary Self-Hosted deployment configuration.

## Product boundary

Self-Hosted Leamout does not provision or administer DIDWW directly.

A self-hosted customer normally connects their own carriers through BYOC. If a self-hosted customer uses Leamout-managed carrier connectivity, they connect to the Leamout-managed SIP service using the SIP configuration supplied by Leamout; they do not receive DIDWW credentials or create DIDWW resources themselves.

Leamout-operated carrier infrastructure owns the DIDWW account and provider-side configuration. A future Backoffice/operator surface should manage that infrastructure.

`LEAMOUT_DEPLOYMENT_ID` identifies a Leamout runtime/deployment. It is not a DIDWW resource identifier and must not be used to create deployment-scoped DIDWW Voice IN trunks.

## Managed number path

```text
customer
  ↓
GET /v1/numbers/available
  ↓
POST /v1/number-orders
  ↓
provider_operations
  ↓
DIDWW order
  ↓
DIDWW DID
  ↓
existing Leamout-managed DIDWW Voice IN trunk
  ↓
Leamout-managed SIP infrastructure
```

The provider-operation executor reconciles by the Leamout provider-operation UUID before purchasing so ambiguous provider responses do not create duplicate orders.

## Existing Leamout model

Use the existing carrier and number primitives.

| Leamout object | DIDWW use |
| --- | --- |
| `carrier_providers` | Built-in DIDWW provider definition. |
| `carrier_connections` | Internal Leamout-managed carrier connection. |
| `carrier_connection_source_ips` | Provider signaling networks accepted for that internal connection. |
| `trunks` | Internal SIP topology for the managed carrier service. |
| `trunk_endpoints` | Internal signaling endpoints. |
| `carrier_connection_provider_resources` | Maps internal carrier connections to provider resources such as a DIDWW `voice_in_trunk`. |
| `phone_numbers` | Purchased DIDs; `provider_resource_id` stores the DIDWW DID resource ID. |
| `voice_bindings` | Maps a DID to a voice application. |

Managed inbound tenancy is derived from the called DID. Provider resources remain internal and do not become customer carrier resources.

## Control-plane adapter

The DIDWW adapter lives under:

```text
server/internal/integrations/carriers/didww/
```

The adapter currently owns provider API behavior needed by the managed-number lifecycle:

- search available DIDs;
- create and reconcile DID orders;
- resolve purchased DIDs;
- attach a purchased DID to an existing DIDWW Voice IN trunk;
- release an owned DID.

It does **not** create or reconcile the DIDWW Voice IN trunk itself. Provider-side carrier infrastructure is an operator concern and should eventually be managed through Backoffice/internal services.

DIDWW API credentials are managed-carrier provider configuration. They do not belong in customer `carrier_connections`.

## Routing purchased numbers

The managed-number executor expects the provider routing target to already exist in Leamout internal state before a customer purchase is submitted.

After a provider order completes, the executor:

1. resolves the purchased DID at DIDWW;
2. reads the snapshotted `voice_in_trunk` provider-resource ID from the provider operation;
3. assigns the DID to that existing DIDWW Voice IN trunk when necessary;
4. completes the managed-number/provider-operation state transactionally.

Provider object IDs remain internal.

If no managed DIDWW routing target has been configured, managed number acquisition must fail before provider purchase rather than inventing provider topology during request handling or process startup.

## Self-Hosted

Self-Hosted deployments must not require:

```text
DIDWW_SOURCE_CIDRS
DIDWW_SIP_ENDPOINTS
SIP_PUBLIC_HOST
SIP_PUBLIC_PORT
SIP_PUBLIC_TRANSPORT
```

They also must not run an internal DIDWW provisioning command.

A self-hosted deployment using Leamout-managed carrier connectivity configures the Leamout-managed SIP service as a carrier connection using the connection details provided by Leamout. DIDWW remains behind the Leamout-managed carrier boundary.

## Capacity and wholesale cost

DIDWW capacity selection is provider state, not customer pricing. The current managed-number acquisition path excludes DID+0 until DIDWW Capacity provisioning is implemented.

`usage_rates` is customer-facing commercial pricing. DIDWW wholesale cost belongs in provider/wholesale accounting resources, not customer rate tables.

## Reconciliation

Provider execution and reconciliation must remain idempotent. Provider state is not authoritative for customer organization ownership or voice bindings; those remain Leamout control-plane state.

Future inventory/routing reconciliation should detect at least:

- DID exists at DIDWW but not in Leamout;
- DID exists in Leamout but no longer exists at DIDWW;
- DIDWW routing no longer points at the expected managed Voice IN trunk;
- provider status changed outside Leamout.

Do not silently delete local state during reconciliation.

## Porting and messaging

Porting remains manual until Leamout owns the full LOA/document/provider-status workflow.

DIDWW SMS/SMPP is outside the managed voice path. When messaging work starts, normalize provider traffic into the Leamout messaging model instead of exposing raw DIDWW payloads.

## Managed voice checklist

- [x] Built-in `didww` carrier provider.
- [x] DIDWW API control-plane configuration.
- [x] Available-number search.
- [x] Durable number orders and provider operations.
- [x] Reconcile-before-purchase provider executor.
- [x] Persist managed DID provider resource IDs.
- [x] Configure purchased DIDs to an existing DIDWW Voice IN trunk.
- [x] Keep provider IDs and credentials internal.
- [x] Remove deployment-scoped DIDWW ingress provisioning from the runtime/deployment surface.
- [ ] Manage Leamout-owned DIDWW carrier infrastructure through Backoffice/internal operator services.
- [ ] Add provider-sandbox acceptance coverage for purchase → route → inbound call.
- [ ] Add broader inventory/routing drift reconciliation.
