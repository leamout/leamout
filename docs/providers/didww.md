# DIDWW provider

DIDWW is the initial managed provider for DID inventory and inbound PSTN routing.

This document records the internal integration contract. DIDWW provider identifiers, credentials, routing resources, and deployment-level carrier topology are not customer-facing API resources.

Hosting mode and carrier mode are independent. See `docs/deployment-carrier-modes.md`. DIDWW belongs to the **Managed Carrier** capability; it is not synonymous with Leamout Cloud and it is not part of BYOC.

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
DIDWW Voice IN trunk
  ↓
deployment managed ingress
  ↓
phone_numbers
```

The provider-operation executor reconciles by the Leamout provider-operation UUID before purchasing so ambiguous provider responses do not create duplicate orders.

## Existing Leamout model

Use the existing carrier and number primitives.

| Leamout object | DIDWW use |
| --- | --- |
| `carrier_providers` | Built-in DIDWW provider definition. |
| `carrier_connections` | Platform-scoped DIDWW managed ingress connection. |
| `carrier_connection_source_ips` | DIDWW signaling networks accepted for that deployment-level connection. |
| `trunks` | Internal deployment-level inbound SIP trunk representing DIDWW ingress topology. |
| `trunk_endpoints` | Optional known DIDWW remote signaling endpoints; never Leamout's own SIP address. |
| `carrier_connection_provider_resources` | Maps the deployment-level connection to DIDWW provider objects such as `voice_in_trunk`. |
| `phone_numbers` | Purchased DIDs; `provider_resource_id` stores the DIDWW DID resource ID. |
| `voice_bindings` | Maps a DID to a voice application. |

Managed inbound tenancy is derived from the called DID. A platform-scoped carrier connection therefore has `organization_id = NULL`; it does not belong to the customer that owns a managed DID.

`scope = 'platform'` means deployment-level/shared internal state. It does **not** mean Leamout Cloud. The same model can represent Managed Carrier ingress for a self-hosted deployment or for a Leamout Cloud deployment.

## Control-plane adapter

The DIDWW adapter lives under:

```text
server/internal/integrations/carriers/didww/
```

Number behavior remains in `numbers.go`, provider routing behavior in `routing.go`, and DIDWW wire/domain types in `model.go`.

DIDWW API credentials are managed-carrier provider configuration. They do not belong in customer `carrier_connections`, which model SIP-facing runtime state.

## Explicit managed-carrier ingress provisioning

DIDWW managed ingress is provisioned by an explicit internal managed-carrier operation. It is not part of API-server or worker startup, and it is not a command in the public self-hosted `leamout` CLI.

Internal deployment/operator automation invokes:

```text
/leamout/internal-provision managed-carrier didww ingress
```

The hosting model does not determine whether this capability is applicable. It is needed when a deployment uses DIDWW-backed Managed Carrier, whether the control plane is self-hosted or Leamout Cloud. BYOC-only deployments do not invoke it.

Provisioning requires a stable deployment identity plus the managed-carrier/provider and public SIP configuration needed to create the DIDWW ingress target:

```text
LEAMOUT_DEPLOYMENT_ID
DIDWW_API_KEY
SIP_PUBLIC_HOST
SIP_PUBLIC_PORT
SIP_PUBLIC_TRANSPORT
DIDWW_SOURCE_CIDRS
```

`DIDWW_SIP_ENDPOINTS` is optional and is only for known provider-side remote signaling endpoints. Entries use `host:port/transport` format.

Credential delivery and entitlement for Self-Hosted + Managed Carrier are separate product concerns from this ingress reconciler. This code must not treat self-hosting itself as ownership of the DIDWW carrier account.

### Provider-side resource

Provisioning reconciles one DIDWW Voice IN trunk using a deployment-scoped external reference:

```text
leamout:<deployment-id>:managed-ingress
```

The desired SIP target uses:

```text
username = +{DID}
host = SIP_PUBLIC_HOST
port = SIP_PUBLIC_PORT
transport = SIP_PUBLIC_TRANSPORT
resolve_ruri = true
auth_enabled = false
```

`+{DID}` makes the called number the user portion of the inbound Request-URI in +E.164 format so the existing OpenSIPS number lookup can resolve the managed `phone_numbers` row.

The resulting DIDWW Voice IN trunk ID is stored only as internal provider metadata:

```text
carrier_connection_provider_resources
  resource_type = voice_in_trunk
  provider_resource_id = <DIDWW Voice IN trunk UUID>
```

### Deployment-level topology

Provisioning transactionally converges:

```text
carrier_provider: didww
        ↓
platform-scoped carrier_connection
  name = DIDWW Managed Ingress
  inbound_enabled = true
  inbound_auth_method = ip
        ↓
platform-scoped inbound trunk
  name = DIDWW Managed Ingress
  direction = inbound
  managed_default = false
        ├── carrier_connection_source_ips
        ├── optional inbound trunk_endpoints
        └── carrier_connection_provider_resources
              voice_in_trunk → DIDWW UUID
```

The inbound trunk is not the managed outbound default. Managed outbound termination remains independent from DIDWW numbering/inbound origination.

The provisioning action is idempotent. Re-running it reconciles the same DIDWW Voice IN trunk and the same deployment-level topology rather than creating duplicates.

## SIP ingress

Expected runtime path:

```text
DIDWW
  ↓ SIP INVITE from configured signaling CIDR
OpenSIPS
  ↓ source-IP carrier resolution
platform-scoped DIDWW carrier_connection
  ↓ called DID
managed phone_numbers row
  ↓ customer organization
voice_bindings
  ↓
voice application
```

`carrier_connection_source_ips` is the runtime source-IP authorization/attribution mechanism. DIDWW signaling CIDRs are deployment data and must not be hard-coded in `opensips.cfg`.

Ingress fails closed when:

- the source IP does not resolve uniquely to a carrier connection;
- the DID is not assigned to that platform connection;
- the managed number is inactive or voice-disabled;
- no voice binding exists for the called number.

## Routing purchased numbers

After a provider order completes, the executor:

1. resolves the purchased DID at DIDWW;
2. reads the snapshotted `voice_in_trunk` mapping from the provider operation;
3. assigns the DID to that DIDWW Voice IN trunk when necessary;
4. creates or reuses the managed `phone_numbers` row;
5. completes the provider operation and customer number order transactionally.

Provider object IDs remain internal; OpenSIPS never needs the DIDWW Voice IN trunk UUID.

If managed ingress has not been explicitly provisioned, managed number acquisition must fail before provider purchase rather than inventing provider topology during request handling or process startup.

## Capacity and wholesale cost

DIDWW capacity selection is provider state, not customer pricing. The current managed-number acquisition path excludes DID+0 until DIDWW Capacity provisioning is implemented.

`usage_rates` is customer-facing commercial pricing. DIDWW wholesale cost belongs in provider/wholesale accounting resources, not customer rate tables.

## Reconciliation

Provider execution and reconciliation must remain idempotent. Provider state is not authoritative for customer organization ownership or voice bindings; those remain Leamout control-plane state.

Future inventory/routing reconciliation should detect at least:

- DID exists at DIDWW but not in Leamout;
- DID exists in Leamout but no longer exists at DIDWW;
- DIDWW routing no longer points at the deployment's expected Voice IN trunk;
- provider status changed outside Leamout.

Do not silently delete local state during reconciliation.

## Porting and messaging

Porting remains manual until Leamout owns the full LOA/document/provider-status workflow.

DIDWW SMS/SMPP is outside the managed voice path. When messaging work starts, normalize provider traffic into the Leamout messaging model instead of exposing raw DIDWW payloads.

## Managed voice checklist

- [x] Built-in `didww` carrier provider.
- [x] DIDWW API deployment configuration.
- [x] Available-number search.
- [x] Durable number orders and provider operations.
- [x] Reconcile-before-purchase provider executor.
- [x] Persist managed DID provider resource IDs.
- [x] Configure purchased DIDs to a DIDWW Voice IN trunk.
- [x] Model deployment-level managed ingress separately from customer BYOC connections.
- [x] Model provider-side Voice IN trunk IDs as internal provider resources.
- [x] Explicitly provision the DIDWW Voice IN trunk and deployment managed-ingress topology.
- [x] Populate platform `carrier_connection_source_ips` from deployment configuration.
- [x] Route managed inbound DIDs through the existing number/binding path.
- [ ] Add provider-sandbox acceptance coverage for purchase → route → inbound call.
- [ ] Define Self-Hosted + Managed Carrier credential/entitlement delivery without exposing provider ownership as BYOC.
- [ ] Add broader inventory/routing drift reconciliation.
