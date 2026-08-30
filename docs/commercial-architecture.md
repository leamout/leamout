# Commercial domain and offline licensing

## Boundaries

The commercial subsystem uses an organization as its account and isolation
boundary. A product groups an offering, a plan describes one purchasable
edition, and a subscription associates an organization with a plan. Provider
identifiers belong to subscription adapters and are not domain identities.

Plans grant default entitlements. Organization entitlements override negotiated
terms and license entitlements override a particular self-hosted license. The
effective value is resolved in this order:

1. license override;
2. organization override;
3. plan default;
4. deny when no value exists.

Entitlements are either Boolean features or non-negative integer limits. Their
keys are stable API identifiers, not display names. An expired entitlement does
not participate in resolution. An explicit disabled feature is different from
an absent feature and therefore overrides a lower-precedence enabled value.

## Hosted and customer-owned enforcement

Hosted services resolve effective entitlements from the control-plane database.
Customer-owned deployments receive the same effective values as a signed
license document. They do not reproduce subscription, billing, or override
resolution locally.

The portable document is a small JSON envelope containing an Ed25519 key ID,
claims payload, and signature. Claims bind the authority to a license and
organization and include issuance, expiration, deployment capacity, and the
effective entitlement snapshot. Private signing keys remain in the control
plane; deployments contain a keyring of public verification keys.

Verification fails closed for an unsupported schema version, unknown key,
invalid signature, malformed entitlement, or expired document. Key rotation is
performed by distributing a keyring containing both old and new public keys,
issuing new documents with the new private key, and removing the old public key
after all valid documents signed by it have expired.

Clock-skew allowance and refresh grace periods are deployment policy, not part
of cryptographic verification. Revocation requires an online refresh or a
short-lived document: an offline verifier cannot discover control-plane state
that was changed after issuance.

## Entitlement resolution and deployment activation

The control plane resolves entitlements into a key-sorted snapshot by applying
the precedence rules above at an explicit evaluation time. An override cannot
change the kind of a lower-precedence entitlement: such data is rejected rather
than interpreted ambiguously.

Deployment activation locks the active, unexpired license row before it checks
capacity. This serializes activations for one license and prevents concurrent
requests from exceeding `max_deployments`. Repeating an active deployment ID is
idempotent; a deactivated ID cannot be reused.

## Active-license document issuance

Document issuance reloads organization-scoped license authority on every
request. Only active, issued, unexpired licenses with a configured signing key
may issue documents. Subscription-backed licenses additionally require an
active subscription and resolve the plan associated with that subscription.
Standalone licenses resolve organization and license entitlements without a
plan layer.

The issuance service depends on a signing-key provider selected by immutable key
ID. The bundled static provider is limited to tests and local development;
production providers should delegate signing to a KMS or HSM and must not expose
private key bytes.

## Next implementation slice

The next slice should add a production KMS-backed provider, PostgreSQL
integration coverage, and authenticated administrative endpoints for activation
and document refresh. Billing-provider adapters should follow those domain
boundaries rather than calling cryptographic primitives directly.
