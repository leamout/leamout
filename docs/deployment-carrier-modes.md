# Deployment and carrier modes

Leamout has two independent axes:

1. **Hosting mode** — where the Leamout control plane runs.
2. **Carrier mode** — who owns and operates the telecom carrier relationship.

Do not infer carrier mode from hosting mode, or hosting mode from carrier mode.

| Hosting mode | Carrier mode | Meaning |
| --- | --- | --- |
| Self-Hosted | BYOC | Customer runs Leamout and connects customer-owned carriers. |
| Leamout Cloud | Managed Carrier | Leamout runs the control plane and provides Leamout-managed carrier connectivity. |
| Self-Hosted | Managed Carrier | Customer runs Leamout while using Leamout-managed carrier connectivity. |
| Leamout Cloud | BYOC | Leamout runs the control plane while the customer connects customer-owned carriers. |

## Model invariants

### Hosting mode

Hosting mode controls deployment ownership and runtime operations. It must not decide whether a carrier connection is BYOC or managed.

- `self-hosted` means the customer operates the Leamout deployment.
- `cloud` means Leamout operates the Leamout deployment.

### Carrier mode

Carrier mode controls carrier ownership and provider-facing operations. It must not decide where Leamout is hosted.

- **BYOC** carrier connections are organization-scoped customer resources.
- **Managed Carrier** ingress/termination resources are deployment-level internal resources used to provide managed telecom service to customer organizations.

Database `scope = 'platform'` means deployment-level/shared internal carrier state. It does **not** mean “Leamout Cloud.” A self-hosted deployment using Managed Carrier may also have platform-scoped managed-carrier resources.

## API boundary

Customer-facing APIs remain the same across hosting modes:

- customer BYOC resources stay organization-scoped;
- managed numbers stay customer-owned number resources;
- provider IDs, provider credentials, wholesale resources, and provider operations remain internal.

Hosting mode must not create separate customer-facing telecom APIs.

## Provisioning boundary

Provider-specific managed-carrier provisioning is an internal capability, not a self-hosted CLI concept and not a cloud-only concept.

A managed-carrier provisioner may be invoked by deployment/operator automation for a deployment that has Managed Carrier enabled. BYOC-only deployments do not invoke it.

For DIDWW, the internal provisioning primitive is:

```text
/leamout/internal-provision managed-carrier didww ingress
```

This command is not part of the public `leamout` self-hosted operator CLI contract. It reconciles deployment-level managed-carrier topology and provider resources; server and worker startup do not run it automatically.

## Design rule

When adding a feature, answer these separately:

```text
Where is Leamout hosted?
  self-hosted | cloud

Who owns the carrier relationship?
  customer/BYOC | Leamout/managed
```

No implementation should collapse those two questions into one mode flag.
