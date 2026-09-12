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
public_key
name
status
activated_at
last_seen_at
deactivated_at
created_at
updated_at
```

Each license has exactly one deployment. `deployment_id` is globally unique, and
`public_key` is the deployment's Ed25519 public identity key.

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

## License binding

A license is bound to one deployment when that deployment is activated. The
database uniqueness constraints prevent a license from being attached to a
second deployment and prevent a deployment identity from being reused by a
different license.

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

Deployment identity is installation generated and stable across normal
restarts. `leamout init` creates an Ed25519 keypair, stores the public key with
the deployment identity, and keeps the private key local. Signed license
artifacts include the public key so copied license material cannot be used by an
installation without the matching private key.
