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

`commercial/licensing` owns durable licenses, license lifecycle, and activated self-hosted installations. It consumes resolved commercial state; it does not decide catalog pricing, subscription lifecycle, or entitlement inheritance.

## Current implementation

The licensing package now uses UUID-backed domain models and SQLC-backed persistence. License and deployment reads/writes are organization-scoped.

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

Creating a row is intentionally not described as cryptographic issuance. A license starts `pending`; it becomes `active` only after trusted licensing-authority code has associated a signing key and completed whatever signed-artifact workflow is introduced by the versioned claims format.

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

## Signed license artifact

The persistent `License` record is **not** the signed license file/token consumed by a self-hosted runtime.

The future signing boundary remains:

```text
commercial state
      ↓
license record
      ↓
versioned claims
      ↓
trusted signer / private key
      ↓
signed artifact
      ↓
self-hosted runtime
      ↓
public-key verification
```

Do not add `signer.go` merely as a scaffold. Add it together with the concrete, versioned claims format and runtime verifier so signing becomes a real compatibility contract.

Useful claims are expected to include Leamout-owned identifiers, validity times, effective feature/limit snapshots, deployment policy, claims version, and signing-key identifier. Exact names/encoding remain deliberately unspecified until that workflow is implemented.

## Key security

- Private signing keys remain in the trusted Leamout licensing authority.
- Self-hosted deployments receive only public verification material and signed commercial claims.
- `signing_key_id` selects verification material and supports key rotation.
- Invalid signatures or malformed claims must fail closed for new commercial actions.
- Runtime verification must not require an external payment provider to be online for every call or media action.

Ed25519 remains a suitable asymmetric signing primitive when the concrete artifact format is implemented.

## Expiry and active calls

Commercial enforcement must not destroy active telecom sessions. If a license expires or becomes invalid while a call is already active, Leamout should not terminate that call solely because of the commercial transition. Enforcement should affect new controlled actions according to policy.

Grace periods, cached renewals, air-gapped licenses, and offline renewal windows remain future policy and are not encoded speculatively.

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
signed license artifact
```

PostgreSQL and Leamout domain state remain authoritative.
