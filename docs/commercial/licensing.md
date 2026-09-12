# Licensing

Licensing carries Leamout commercial authority into self-hosted deployments without depending on Cloud PAYG or a customer subscription lifecycle.

Self-Hosted software licensing is the sole exception to Leamout's prepaid pay-as-you-go commercial model.

The enterprise software license is sold separately from Cloud and managed usage, normally through a negotiated agreement, invoice, and bank transfer. Paying for a Self-Hosted license does not create wallet balance or managed-usage credit.

## Canonical rule

One Self-Hosted license authorizes exactly one Leamout deployment.

All Self-Hosted capabilities available in that Leamout release are included. Licensing is not a feature, entitlement, edition, or capacity policy system.

```text
one license
    ↓
one deployment
```

A customer operating production, disaster-recovery, and staging as three independently operated Leamout environments needs three licenses.

A deployment may contain multiple API, SIP, media, worker, database, or coordination nodes. Those nodes do not consume separate licenses when they belong to the same deployment identity.

## Boundary

```text
enterprise agreement
        ↓
Self-Hosted license
        ↓
one deployment identity
        ↓
signed deployment-bound license artifact
```

`commercial/licensing` owns durable licenses, one-to-one deployment activation, and the signed artifact protocol consumed by self-hosted runtimes.

Licensing does not own Cloud wallet balance, wallet funding, managed-carrier authorization, customer subscriptions, feature entitlements, capacity limits, or postpaid usage settlement.

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

The durable license represents permission for one deployment to operate Leamout.

Conceptually it contains:

```text
id
organization_id
status
signing_key_id
issued_at
expires_at
created_at
updated_at
```

There is no `subscription_id`, `max_deployments`, feature list, or limit list.

A newly created license starts as `pending`. Activating it binds the license to one deployment identity. An already-expired license cannot be activated.

## Deployment identity

`leamout init` creates one durable deployment identity for the self-hosted environment.

The deployment identity consists of:

```text
deployment_id
deployment keypair
```

The deployment private key remains local and must never be sent to Leamout Cloud. The corresponding public key is used during activation so Leamout can bind the license to that deployment.

Do not derive deployment identity from MAC address, IP address, hostname, container ID, disk serial number, or other unstable hardware/runtime properties.

## Activation

Activation binds an unused license to one deployment.

```text
activation credential
        ↓
deployment_id + deployment public key
        ↓
Leamout validates license
        ↓
license becomes bound to deployment
        ↓
Leamout issues signed license artifact
```

Once bound, the same license cannot activate a different deployment unless an explicit re-host/transfer workflow first retires the old binding.

Activation retries for the same deployment identity should be idempotent.

## Signed license protocol

The runtime consumes a signed, deployment-bound artifact rather than the database row directly.

The signed claims should carry only licensing identity and validity information:

```text
license_id
organization_id
deployment_id
deployment_public_key
issued_at
expires_at
```

There are no feature or capacity claims.

The artifact is signed by Leamout using a trusted signing key and verified locally using public verification material shipped through the trusted release channel.

## Protection against copying

Copying only the signed license artifact must not be sufficient to operate another deployment.

Runtime authorization should require both:

```text
valid signed Leamout license artifact
        +
proof of possession of the bound deployment private key
```

The private key belongs to the deployment and is not contained in the license artifact.

This makes ordinary license sharing fail without tying the license to replaceable hardware.

A full malicious clone of both deployment state and private key cannot be made impossible when the customer controls the host. Leamout may detect suspicious duplicate use through periodic authenticated check-ins, while contractual enforcement remains part of the enterprise license relationship.

## Offline operation

Normal self-hosted call handling must not require continuous Leamout Cloud availability.

After activation, the runtime verifies signed license material locally. License artifacts should have finite validity and may be periodically renewed, but transient Cloud outages must not immediately disable otherwise valid communications workloads.

## Lifecycle

```text
pending ─────→ active ─────→ suspended ─────→ active
   │             │              │
   │             ├──────────────┴────→ expired
   │             └───────────────────→ revoked
   └─────────────────────────────────→ revoked

expired / revoked = terminal
```

License changes govern admission of new controlled work. They must not intentionally terminate an already-established telecom session solely because the license state changes or an artifact expires mid-session.

## Re-hosting

A replacement host or environment must use an explicit license transfer process rather than silently creating a second active deployment.

Conceptually:

```text
old deployment binding
        ↓
retire / transfer
        ↓
new deployment identity
        ↓
activate same commercial license
```

At no time should one license authorize two independently operated deployment identities.

## Key rotation

Leamout signs artifacts with a private signing key identified by `key_id`. Self-hosted runtimes receive only public verification material.

During signing-key rotation, old and new public keys may overlap so already-issued artifacts remain verifiable until their normal expiry. Leamout private signing keys must never be shipped to self-hosted deployments.

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

PostgreSQL and Leamout licensing state remain authoritative for self-hosted commercial authority.
