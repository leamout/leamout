# Licensing

Licensing carries Leamout commercial authorization into self-hosted deployments without exposing private signing capability to the self-hosted runtime.

## Commercial flow

```text
organization
    ↓
subscription
    ↓
effective entitlements
    ↓
license
    ↓
deployment(s)
```

## Current license model

```text
id
organization_id
subscription_id
status
max_deployments
signing_key_id
issued_at
expires_at
created_at
updated_at
```

Current states are:

```text
pending
active
suspended
expired
revoked
```

The database validates state names and basic temporal/deployment limits. Service logic owns valid transitions and issuance policy.

## Package boundary

License verification and license authority must remain separate.

```text
server/internal/commercial/licensing
        │
        │ private signing authority
        ▼
     signed license
        │
        │ public verification
        ▼
server/pkg/license
```

`server/pkg/license` may contain a public-safe format, verifier, and verification errors. It must not contain private signing keys or functionality that allows a public/self-hosted binary to generate valid commercial licenses.

Private signing behavior belongs under `server/internal/commercial/licensing`.

## Intended operations

Licensing service behavior is expected to include operations such as:

```text
issue
activate
renew
deactivate
suspend
expire
revoke
```

These are service methods/use cases, not separate files merely because they may have separate API endpoints.

## Claims

A signed self-hosted license should be based on Leamout-owned commercial identifiers and effective entitlements. Useful claims can include:

```text
organization
license identifier
deployment identifier or deployment policy
issued/expiry times
features
limits
max deployments
signing key identifier
```

The exact signed format should be versioned before it becomes a compatibility contract.

## Key security

- Private signing keys remain in the private commercial authority.
- Self-hosted binaries receive only verification material.
- `signing_key_id` allows key rotation and verification-key selection.
- License verification must fail closed for invalid signatures or malformed claims.

Ed25519 is suitable for this asymmetric signing boundary.

## Expiry and availability

Commercial enforcement must not destroy active telecom sessions.

If a license expires or becomes invalid while a call is already active, Leamout should not terminate that call solely because of the commercial transition. Enforcement should generally affect new chargeable/controlled actions according to policy.

Self-hosted licensing may later support cached renewal windows, grace periods, and offline/air-gapped licenses. Those policies are not encoded in the current schema.

## Relationship to payment providers

Never implement:

```text
provider webhook
    ↓
direct license signing
```

Use:

```text
provider event
    ↓
payment/subscription reconciliation
    ↓
entitlement resolution
    ↓
licensing service
    ↓
signed license
```

This keeps license issuance provider-independent and auditable through Leamout domain state.
