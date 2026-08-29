# Deployments

A deployment represents one activated self-hosted Leamout installation under a commercial license.

## Model

```text
organization
    ↓
license
    ↓
deployment
```

Deployments are license-scoped rather than directly organization-scoped in the schema. Tenant ownership is derived through the license.

## Current fields

```text
license_id
deployment_id
name
status
activated_at
last_seen_at
deactivated_at
created_at
updated_at
```

`deployment_id` is unique within a license.

## Lifecycle

```text
activation
    ↓
 active
    │
    ├── heartbeat / last_seen_at
    │
    └── deactivation
            ↓
       deactivated
```

The current schema does not reactivate a deactivated row implicitly. Reactivation policy should be explicit in the licensing/deployment service.

## Deployment limit

Each license carries `max_deployments`.

The intended invariant is:

```text
active deployments for license <= license.max_deployments
```

A simple `COUNT` followed by `INSERT` is vulnerable to a race when two activations happen concurrently. The service must enforce the limit transactionally, for example by locking the license row before counting and inserting or by using another database serialization strategy.

The SQL query layer provides ownership-aware lookup/count/create primitives; the service owns the transaction boundary.

## Heartbeats

`last_seen_at` records recent contact from an active deployment.

```text
deployment runtime
      ↓
renew/heartbeat request
      ↓
verify organization + license + deployment
      ↓
update last_seen_at
```

A heartbeat is operational evidence, not proof that a license should be renewed. Renewal remains a licensing/commercial decision.

## Security

A deployment ID is not sufficient authorization. Queries must resolve:

```text
organization_id
    ↓
license belongs to organization
    ↓
deployment belongs to license
```

The organization must also be active and not deleted.

## Identity

Deployment identity should be installation/application generated and stable across normal restarts. Avoid brittle hardware fingerprinting as the primary license identity mechanism.
