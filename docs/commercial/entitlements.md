# Entitlements

Entitlements describe commercial capabilities and limits independently of payment-provider implementation.

## Module boundary

```text
catalog + subscriptions
          ↓
     entitlements
          ↓
        state
          ↓
      licensing
```

`commercial/entitlements` owns durable feature/limit grants and effective entitlement resolution. It does not own subscription lifecycle, runtime policy caching, license signing, usage metering, rating, invoicing, or payments.

## Scopes

An entitlement belongs to exactly one scope:

```text
plan
organization
license
```

The database enforces this with an owner constraint: exactly one of `plan_id`, `organization_id`, or `license_id` must be non-null.

Plan entitlements are reusable defaults. Organization entitlements are overrides or additional grants for one organization. License entitlements are the durable snapshot issued to a specific self-hosted license.

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

The service validates the same value-shape rule enforced by PostgreSQL:

```text
feature → enabled required, limit_value absent
limit   → limit_value >= 0, enabled absent
```

Entitlement keys are trimmed, non-empty stable machine identifiers and cannot contain whitespace.

## Organization resolution

Effective organization entitlements require the organization's current subscription. The subscription identifies the plan whose entitlement defaults apply.

```text
current subscription
        ↓
       plan
        ↓
plan entitlements
        ↓
organization overrides
        ↓
effective organization entitlement set
```

For the same entitlement key, an effective organization entitlement overrides the plan value. Organization-only keys are also allowed.

An override must retain the inherited key's kind. A feature cannot override a plan limit with the same key, and a limit cannot override a plan feature with the same key.

If the organization has no current `active` or `past_due` subscription, effective organization resolution fails rather than returning an empty capability set.

## License resolution

License entitlements are resolved independently from current plan or organization state:

```text
license entitlement snapshot
        ↓
resolved license entitlement set
```

This preserves the commercial state that was issued to a self-hosted license. Changing a plan or organization entitlement later does not implicitly rewrite an existing license snapshot.

## Time bounds

Entitlements can optionally have `starts_at` and `expires_at` values.

A durable list operation returns the stored records regardless of whether they are current, future, or expired. Effective resolution applies the evaluation time:

```text
starts_at <= evaluation time
expires_at >= evaluation time
```

when those bounds exist.

The service exposes deterministic `...At` resolution methods for tests and workflows that must evaluate commercial state at a specific time.

## Uniqueness

Within each scope, an entitlement key is unique:

```text
(plan_id, entitlement_key)
(organization_id, entitlement_key)
(license_id, entitlement_key)
```

This keeps override resolution deterministic within one scope.

## Persistence

All entitlement persistence goes through SQLC-generated methods backed by `server/internal/database/queries/entitlements.sql`.

`commercial/entitlements/repository.go` contains no SQL strings and performs no direct `Query`, `QueryRow`, or `Exec` persistence calls. It only adapts SQLC rows/errors to entitlement domain types.

Database writes repeat scope validity checks:

- plan grants require an active plan whose parent product is active;
- organization grants require an active, non-deleted organization;
- license grants and reads verify the license belongs to the requested organization.

## What entitlements do not represent

Entitlements are not payment records, usage ledgers, pricing rows, or runtime counters. They answer questions such as:

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
