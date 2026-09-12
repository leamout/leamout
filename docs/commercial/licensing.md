# Licensing

Licensing carries Leamout commercial authority into self-hosted deployments without depending on Cloud PAYG or a customer subscription lifecycle.

Self-Hosted software licensing is the sole exception to Leamout's prepaid pay-as-you-go commercial model.

The enterprise software license is sold separately from Cloud and managed usage, normally through a negotiated agreement, invoice, and bank transfer. Paying for a Self-Hosted license does not create wallet balance or managed-usage credit.

## Boundary

```text
enterprise agreement
        ↓
Self-Hosted license
        ↓
deployment(s)
        ↓
signed deployment artifact
```

`commercial/licensing` owns durable licenses, license lifecycle, activated self-hosted installations, and the signed artifact protocol consumed by self-hosted runtimes.

Licensing does not own Cloud wallet balance, wallet funding, managed-carrier authorization, customer subscriptions, or postpaid usage settlement.

## Commercial combinations

```text
Self-Hosted + BYOC
        ↓
enterprise Self-Hosted software license
```

```text
Self-Hosted + Managed
        ↓
enterprise Self-Hosted software license
        +
prepaid managed-usage wallet
```

A Self-Hosted + Managed customer must satisfy both independently. An active license does not authorize Leamout to incur managed-provider obligations without prepaid wallet funds.

## License model

A license contains:

```text
id
organization_id
status
max_deployments
signing_key_id
issued_at
expires_at
created_at
updated_at
```

There is no `subscription_id` dependency.

License creation accepts explicit self-hosted licensing authority, including `max_deployments`, optional signing key identity, and optional expiration.

A newly created license starts as `pending`. Activating a license requires a signing key. An already-expired license cannot be activated.

## Lifecycle

```text
pending ─────→ active ─────→ suspended ─────→ active
   │             │              │
   │             ├──────────────┴────→ expired
   │             └───────────────────→ revoked
   └─────────────────────────────────→ revoked

expired / revoked = terminal
```

Repeated transitions to the current state are treated idempotently where supported by the service/repository lifecycle rules.

## Deployment activation

A deployment is one self-hosted Leamout installation using a license slot.

```text
license
   ├── deployment node-01
   ├── deployment node-02
   └── deployment node-03
```

Activation requires:

```text
license.status = active
license not expired
active deployments < max_deployments
```

The same active `deployment_id` is idempotent and does not consume another slot. A deactivated deployment identity is not silently reactivated.

The repository protects deployment limits transactionally so concurrent activations cannot exceed `max_deployments`.

## Signed license protocol v1

The persistent `License` row is not the artifact consumed by a self-hosted runtime.

Version 1 uses Ed25519 signatures over versioned, deployment-bound claims.

```text
license authority
      ↓
LicenseClaimsV1
      ↓
canonical payload
      ↓
Ed25519 signature
      ↓
versioned envelope
      ↓
self-hosted runtime
      ↓
public-key verification
```

### Deployment binding

A signed artifact is bound to one activated `deployment_id`.

An artifact issued for one deployment must fail verification when presented by another deployment. This prevents one artifact from being copied across unlimited installations and bypassing `max_deployments`.

### Claims v1

Current claims carry:

```text
license_id
organization_id
deployment_id
issued_at
expires_at
features[]
limits[]
```

There is no `subscription_id` claim.

Feature and limit claims are part of the signed license protocol. They are not the deleted `commercial/entitlements` service or database table.

Claim keys are normalized and sorted before encoding. A key cannot appear as both a feature and a limit. Limit values must be non-negative. Times are normalized to UTC whole seconds so canonical payloads remain deterministic.

### Envelope v1

The transport envelope carries:

```text
version   = 1
algorithm = Ed25519
key_id
payload   = base64url(canonical claims JSON)
signature = base64url(Ed25519 signature)
```

Signatures use domain separation and authenticate both key identity and payload.

The decoder rejects unsupported versions/algorithms, malformed payloads, unknown key IDs, invalid signatures, deployment mismatches, not-yet-valid artifacts, and expired artifacts.

## Key rotation

The trusted authority signs with a private key selected by `key_id`. Self-hosted runtimes receive only the public-key keyring.

During rotation, old and new public keys may overlap so already-issued artifacts remain verifiable until their normal expiry. Private signing keys must never be shipped to self-hosted deployments.

## Offline validity and revocation

A disconnected self-hosted runtime cannot learn immediately that server-side state changed.

Signed artifacts therefore carry their own finite `expires_at` validity boundary. Artifact validity may be shorter than the durable license expiration so the runtime periodically refreshes trusted authority without pretending fully offline instant revocation exists.

## Expiry and active telecom sessions

Commercial transitions should govern new controlled actions. They should not destroy an already-active telecom session solely because a license artifact expires mid-session.

Exact grace, refresh, air-gapped, and offline-renewal policy remains above the cryptographic protocol.

## Payment and provider independence

Self-Hosted software license settlement is an enterprise procurement concern, not a wallet top-up workflow.

Payment-provider events must never directly sign licenses.

Managed-provider authorization must never be inferred from license state.

```text
enterprise license settlement
        ↓
license lifecycle

prepaid wallet funding
        ↓
managed usage authorization
```

These paths remain separate even when the same organization uses both.

PostgreSQL and Leamout licensing state remain authoritative for self-hosted authority.
