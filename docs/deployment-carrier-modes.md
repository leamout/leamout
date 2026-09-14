# Deployment and carrier modes

Leamout has two runtime placements, but they do not expose the same connectivity choices.

```text
                         LEAMOUT
                            │
              ┌─────────────┴─────────────┐
              │                           │
         LEAMOUT CLOUD                SELF-HOSTED
              │                           │
        ┌─────┴─────┐                     │
        │           │                     │
       BYOC       MANAGED                BYOC
        │           │                     │
 customer        Leamout             customer selects
 selects         supplies            the carrier
 carrier         telecom                  │
                                      any carrier,
                                      including Leamout
```

This produces three product delivery modes:

| Runtime | Connectivity | Meaning |
| --- | --- | --- |
| Self-Hosted | BYOC | Customer runs Leamout and selects/configures the carrier, including Leamout Carrier if desired. |
| Leamout Cloud | BYOC | Leamout runs the runtime while the customer selects/configures the carrier. |
| Leamout Cloud | Managed | Leamout runs the runtime and supplies/selects the telecom connectivity. |

There is no separate **Self-Hosted + Managed Carrier** delivery mode.

## Definitions

### Runtime placement

Runtime placement answers where the Leamout communications runtime executes.

- `self-hosted` means the customer operates the Leamout runtime on customer-controlled infrastructure.
- `cloud` means Leamout operates the runtime.

### BYOC

BYOC means the customer selects and configures the carrier connection used by the runtime.

The carrier does not have to be owned by the customer. It may be any supported carrier or telecom provider, including Leamout Carrier.

For example, all of these are ordinary Self-Hosted + BYOC connections:

```text
Self-Hosted Leamout
        │
        ├── DIDWW
        ├── CommPeak
        ├── Telnyx
        ├── local carrier
        └── Leamout Carrier
```

Using Leamout Carrier does not change the deployment mode. In that relationship Leamout is acting as a telecom provider, while the customer still chooses and configures the carrier connection in their self-hosted runtime.

### Managed

Managed connectivity exists only in Leamout Cloud.

Managed means Leamout supplies, selects, operates, and authorizes the telecom connectivity behind the Cloud runtime. The customer does not configure Leamout's wholesale carrier connections or physical SIP endpoints.

Leamout may fulfill managed connectivity through DIDWW, CommPeak, other wholesalers, or future direct interconnects. Those upstream details remain behind the Leamout-operated boundary.

## Resource ownership

Carrier scope, not provider brand, determines the resource boundary.

```text
organization-scoped carrier connection
        = customer-selected connectivity / BYOC

platform-scoped carrier connection
        = Leamout-owned upstream connectivity for Cloud Managed
```

An organization-scoped carrier connection may use `provider = leamout`. It is still BYOC.

A platform-scoped carrier connection is internal Leamout infrastructure. It is not exposed as a customer's carrier connection.

Trunks and endpoints inherit that ownership boundary from their carrier connection. A physical trunk endpoint does not need its own `byoc` or `managed` type.

## API boundary

Customer-facing telecom APIs should preserve the same carrier primitives across runtimes while respecting the runtime's allowed connectivity model:

- organization-scoped carrier connections are customer-selected/BYOC resources;
- self-hosted exposes organization-scoped carrier connectivity only;
- Cloud may expose organization-scoped BYOC resources and Cloud-managed product workflows;
- platform-scoped carrier connections, provider credentials, wholesale resources, and physical managed endpoints remain internal;
- managed numbers remain organization-owned Leamout number resources even though their upstream provider resources remain internal.

Do not create a special self-hosted telecom API merely because the selected provider is Leamout.

## Provider administration boundary

Provider-specific wholesale administration is a Leamout operator concern.

A self-hosted customer choosing Leamout Carrier receives customer-facing carrier connection details in the same way they would configure another supported carrier. They do not receive DIDWW, CommPeak, or other Leamout wholesale credentials.

Cloud Managed may use platform-scoped provider resources internally. Runtime/server startup must not create or reconcile wholesale provider topology as a side effect.

`LEAMOUT_DEPLOYMENT_ID` identifies a Leamout runtime/deployment. It is not a provider resource identifier.

## Design rule

When adding a telecom feature, determine the runtime first:

```text
Self-Hosted
    → connectivity is BYOC
    → customer chooses the carrier
    → provider may be Leamout

Leamout Cloud
    → BYOC
      or
    → Managed
```

Then determine resource ownership:

```text
customer-selected carrier
    → organization-scoped carrier connection

Leamout Cloud managed upstream
    → platform-scoped carrier connection
```

Do not infer connectivity ownership from the carrier's brand. `provider = leamout` does not mean `managed` when the connection is organization-scoped.
