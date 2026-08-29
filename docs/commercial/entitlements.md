# Entitlements

Entitlements describe commercial capabilities and limits independently of payment-provider implementation.

## Scopes

An entitlement belongs to exactly one scope:

```text
plan
organization
license
```

The database enforces this with an owner constraint: exactly one of `plan_id`, `organization_id`, or `license_id` must be non-null.

## Kinds

Current entitlement kinds are:

```text
feature
limit
```

A feature uses `enabled`:

```text
voice.enabled = true
recording.enabled = false
```

A limit uses `limit_value`:

```text
max.deployments = 3
max.concurrent.calls = 100
```

The database prevents a feature from also carrying a limit value and prevents a limit from carrying a feature boolean.

## Resolution model

The intended resolution order is:

```text
plan entitlements
        ↓
organization overrides
        ↓
effective organization entitlement set
        ↓
license issuance
        ↓
license entitlement snapshot
```

Plan entitlements define reusable defaults. Organization entitlements allow negotiated or operational overrides without requiring a new plan. License entitlements preserve the effective capabilities issued to a self-hosted license.

## Time bounds

Entitlements can optionally have `starts_at` and `expires_at` values. A service evaluating an entitlement should consider the evaluation time in addition to its scope and value.

## Uniqueness

Within each scope, an entitlement key is unique:

```text
(plan_id, entitlement_key)
(organization_id, entitlement_key)
(license_id, entitlement_key)
```

This keeps override resolution deterministic within one scope.

## Security

Organization entitlements require an active, non-deleted organization. License entitlements must resolve the license through the requested organization rather than trusting a license ID alone.

See [security.md](security.md).

## What entitlements do not represent

Entitlements are not payment records, usage ledgers, or pricing rows. They answer questions such as:

```text
Is feature X enabled?
What is limit Y?
What capabilities were issued into this license?
```

They should not answer:

```text
How much did this call cost?
Was this invoice paid?
How many seconds were consumed this month?
```
