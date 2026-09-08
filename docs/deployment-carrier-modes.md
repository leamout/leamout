# Deployment and carrier modes

Leamout has two independent axes:

1. **Hosting mode** — where the Leamout control plane runs.
2. **Carrier mode** — who owns and operates the telecom carrier relationship.

Do not infer carrier mode from hosting mode, or hosting mode from carrier mode.

| Hosting mode | Carrier mode | Meaning |
| --- | --- | --- |
| Self-Hosted | BYOC | Customer runs Leamout and connects customer-owned carriers. |
| Leamout Cloud | Managed Carrier | Leamout runs the control plane and provides Leamout-managed carrier connectivity. |
| Self-Hosted | Managed Carrier | Customer runs Leamout and connects their deployment to Leamout's managed SIP service using connection details supplied by Leamout. |
| Leamout Cloud | BYOC | Leamout runs the control plane while the customer connects customer-owned carriers. |

The routing rule is two-dimensional:

```text
Runtime = where Leamout executes
Connectivity = whose carrier network Leamout uses
```

## Model invariants

### Hosting mode

Hosting mode controls deployment ownership and runtime operations. It must not decide whether a carrier connection is BYOC or managed.

- `self-hosted` means the customer operates the Leamout deployment.
- `cloud` means Leamout operates the Leamout deployment.

### Carrier mode

Carrier mode controls carrier ownership and provider-facing operations. It must not decide where Leamout is hosted.

- **BYOC** means the customer owns and configures the upstream carrier relationship.
- **Managed Carrier** means Leamout owns and operates the upstream carrier/provider relationship.

A self-hosted customer using Managed Carrier does not receive or configure Leamout's wholesale provider credentials. They configure the Leamout-managed SIP service as a carrier connection on their own deployment.

Provider-specific resources such as DIDWW Voice IN trunks, CommPeak credentials, provider source networks, and wholesale resource IDs remain on the Leamout-operated side of that boundary.

## API boundary

Customer-facing APIs remain consistent across hosting modes:

- customer BYOC resources stay organization-scoped;
- managed numbers stay customer-owned number resources;
- provider IDs, provider credentials, wholesale resources, and provider operations remain internal.

Hosting mode must not expose provider ownership details through separate customer-facing telecom APIs.

## Provider administration boundary

Provider-specific managed-carrier administration is a Leamout operator concern, not a deployment/bootstrap concern.

Self-hosted deployment configuration must not require DIDWW or CommPeak provider-infrastructure settings merely because the customer elects to use Leamout-managed carrier service.

Leamout-owned provider resources should be managed through internal services and the future Backoffice operator surface. Runtime/server startup must not create or reconcile wholesale provider topology as a side effect.

`LEAMOUT_DEPLOYMENT_ID` identifies a Leamout runtime/deployment. It is not a provider resource identifier.

## Design rule

When adding a feature, answer these separately:

```text
Where is Leamout hosted?
  self-hosted | cloud

Who owns the carrier relationship?
  customer/BYOC | Leamout/managed
```

If the carrier is Leamout-managed, provider credentials and provider-side topology stay behind the Leamout-managed boundary. If the runtime is self-hosted, that deployment connects to the managed SIP service using the customer-facing SIP connection details supplied by Leamout.
