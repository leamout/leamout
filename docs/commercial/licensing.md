# Licensing

Licensing carries Leamout commercial authorization into self-hosted deployments without exposing private signing capability to the self-hosted runtime.

## Boundary

```text
catalog
   ↓
subscriptions
   ↓
entitlements
   ↓
state
   ↓
licensing
   ↓
self-hosted deployment(s)
```

`commercial/licensing` owns durable licenses, license lifecycle, activated self-hosted installations, and the signed artifact protocol consumed by self-hosted runtimes. It consumes resolved commercial state; it does not decide catalog pricing, subscription lifecycle, or entitlement inheritance.

## Current implementation

The licensing package uses UUID-backed domain models and SQLC-backed persistence. License and deployment reads/writes are organization-scoped.

A new license is created from current commercial state:

```text
current active subscription
        +
effective organization entitlements
        ↓
max.deployments
        ↓
pending license
```

`max_deployments` is a durable snapshot of the resolved `max.deployments` entitlement at license creation time. A caller cannot choose a larger deployment limit directly.

Creating a row is not cryptographic issuance. A license starts `pending`; it becomes `active` only after trusted licensing-authority code has associated a signing key and is ready to issue deployment-bound artifacts.

## License model

```text
id                 UUID
organization_id    UUID
subscription_id    UUID
status
max_deployments
signing_key_id
issued_at
expires_at
created_at
updated_at
```

Lifecycle:

```text
pending ─────→ active ─────→ suspended ─────→ active
   │             │              │
   │             ├──────────────┴────→ expired
   │             └───────────────────→ revoked
   └─────────────────────────────────→ revoked

expired / revoked = terminal
```

Repeated transitions to the current state are idempotent. License transitions use serializable database transactions so concurrent lifecycle changes cannot silently overwrite one another.

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

The same active `deployment_id` is idempotent and returns the existing deployment rather than consuming another slot. A deactivated deployment identity is not silently reactivated.

The repository performs the read/count/create sequence in a PostgreSQL `SERIALIZABLE` transaction and retries serialization failures. That protects `max_deployments` from concurrent activation races without embedding raw SQL in the commercial repository.

Deployment lifecycle currently supports:

```text
activate
list
touch / last-seen
deactivate
```

## Signed license protocol v1

The persistent `License` record is **not** the signed artifact consumed by a self-hosted runtime. Version 1 now defines a concrete Ed25519 protocol.

```text
trusted license claims
        ↓
canonical v1 payload
        ↓
Ed25519 signer
        ↓
versioned envelope
        ↓
self-hosted runtime
        ↓
public-key keyring
        ↓
local verification
```

### Deployment binding

A signed artifact is bound to one activated `deployment_id`.

```text
license L
  ├── node-01 → artifact for node-01
  └── node-02 → artifact for node-02
```

An artifact issued for `node-01` fails verification when presented by `node-02`. This is required because a reusable license-wide token could otherwise be copied to unlimited machines and bypass `max_deployments` while offline.

### Claims v1

`LicenseClaimsV1` carries:

```text
license_id
organization_id
subscription_id
deployment_id
issued_at
expires_at
features[]
limits[]
```

Feature and limit keys are normalized and sorted before encoding. A key cannot appear as both a feature and a limit. Times are normalized to UTC whole seconds. That makes the v1 payload deterministic for the same normalized claims.

Features and limits use sorted arrays in the wire format rather than JSON maps. The runtime reconstructs maps after verification.

### Envelope v1

The transport envelope carries:

```text
version   = 1
algorithm = Ed25519
key_id
payload   = base64url(canonical claims JSON)
signature = base64url(Ed25519 signature)
```

Signatures authenticate both the selected key identity and payload using domain separation:

```text
"leamout-license-v1\0" || key_id || "\0" || payload
```

This prevents `key_id` substitution even if the same public key is temporarily registered under multiple rotation identifiers, and prevents a payload from another protocol from being treated as a Leamout license artifact.

The decoder rejects unknown JSON fields, unsupported versions, unsupported algorithms, malformed base64, malformed claims, invalid signatures, unknown key IDs, wrong deployment IDs, artifacts used before `issued_at`, and artifacts at or after `expires_at`.

## Key rotation

The trusted authority signs with a private Ed25519 key selected by `key_id`. Self-hosted runtimes receive a public-key keyring:

```text
key-old → old public key
key-new → new public key
```

During rotation both public keys may remain in the runtime keyring so already-issued artifacts continue to verify until normal expiry. Once no valid artifact depends on `key-old`, that public key can be removed.

Private signing keys must never be shipped to self-hosted deployments.

## Offline validity and revocation

Offline verification has an unavoidable availability/security tradeoff: a self-hosted runtime cannot learn that a server-side license was suspended or revoked while it is disconnected.

Therefore the signed artifact has its own finite `expires_at` validity boundary. The licensing authority can issue an artifact whose lifetime is shorter than the durable license expiration:

```text
durable license expires in 1 year
        ↓
signed deployment artifact valid for N hours/days
        ↓
periodic trusted refresh
```

The exact online refresh/grace policy remains a product decision. The cryptographic protocol does not pretend instant revocation exists for fully offline installations.

## Entitlement snapshot rule

Signed feature/limit claims must come from a trusted, durable license entitlement snapshot. They must not be assembled from arbitrary request payloads.

The current license creation flow durably snapshots `max.deployments`, but full feature/limit snapshot persistence is not yet atomic with license creation. Until that transaction boundary is implemented, the signer protocol remains a cryptographic primitive and should not be wired to an endpoint that signs mutable organization state opportunistically.

The intended issuance flow is:

```text
resolved commercial state
        ↓
create license + durable entitlement snapshot atomically
        ↓
activate deployment
        ↓
load license snapshot
        ↓
build LicenseClaimsV1
        ↓
sign with license.signing_key_id
        ↓
return deployment-bound artifact
```

## Expiry and active calls

Commercial enforcement must not destroy active telecom sessions. If a signed artifact expires while a call is already active, Leamout should not terminate that call solely because of the commercial transition. Enforcement should affect new controlled actions according to policy.

Grace periods, cached renewals, air-gapped licenses, and offline renewal windows remain future policy above the cryptographic verification layer.

## Provider independence

Never implement:

```text
payment provider webhook
        ↓
direct license signing
```

Use:

```text
provider event
     ↓
Leamout subscription/payment reconciliation
     ↓
commercial state
     ↓
licensing authority
     ↓
durable license snapshot
     ↓
signed deployment artifact
```

PostgreSQL and Leamout domain state remain authoritative.
