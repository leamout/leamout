# Open-core editions and identity

Leamout is one platform with **two deployments** (Self-Hosted, Cloud) and
**two editions** (Community, optional Enterprise). Deployment describes who
operates the software; edition describes which separately distributed
capabilities and support are included. Neither changes the SIP/audio hot path.

## Community (Apache-2.0)

The public repository is Apache-2.0. Community Self-Hosted works without
a Leamout license file, vendor login, online license verification, or Cloud
subscription. Community includes:

- SIP/BYOC, AI provider integrations/BYOA, S3-compatible storage/BYOS.
- Voice Agents, realtime audio, tool gateway, agent control and observability.
- Secure email/password and email-OTP authentication, password recovery,
  user and session management.
- Organization isolation and basic Owner/Admin/Member permissions, API keys
  and essential security/operational logs.
- Docker deployment using the Community code and user-managed infrastructure.

Security patches, tenant isolation, account recovery and core observability
must **never** be disabled because a commercial entitlement is missing.

## Enterprise (separately licensed extensions)

Potential Enterprise offerings: advanced SAML/OIDC federation and identity
policy, SCIM provisioning, custom RBAC, tamper-evident audit exports and
long retention, compliance tools, supported air-gapped/HA/VPC distribution,
premium templates, and contractual support. These are product boundaries,
not claims that every feature is already shipped as proprietary code.

The existing repository already contains some SSO, SCIM, retention and
networking code under Apache-2.0. Its Apache rights remain unchanged. A
commercial offering must provide **new** separately distributed code,
packaging, services and/or support; runtime flags alone do not turn the public
implementation proprietary. Audit any proposed split before publishing.

### Signed license semantics

Commercial Self-Hosted uses a read-only, Ed25519-signed license to enable
optional Enterprise extensions in offline installations. No license file means
`community` edition. Missing, expired, invalid or untrusted Enterprise
licenses must disable optional Enterprise features, not stop Community calls.
Cloud capabilities remain managed by plans and organization entitlements.

The license issuer's private key never ships with Community or Enterprise
runtime binaries. Future Enterprise extensions must verify relevant claims
through a narrow capability API and independently enforce tenant authorization.

## Authentication boundary

Authentication and edition are **orthogonal**. Community always includes
secure login; Enterprise may add enterprise federation and identity controls.

- Community Self-Hosted: built-in verified email/password/OTP for now.
- Cloud: built-in login for now; Auth0 may later act as a managed OIDC provider.
- Self-Hosted with customer identity provider: future standards-based OIDC
  integration, with Keycloak or another customer-managed provider.
- Enterprise: contractual advanced identity federation, SCIM and governance.

No Auth0 or Keycloak dependency is added merely by defining this model.
Leamout keeps local user IDs, organization memberships, authorization,
application API credentials, Polar Cloud subscriptions and commercial
entitlements independently of the provider used for sign-in.

## Commercial responsibility

- Community Self-Hosted: free under Apache-2.0; customer pays own infra,
  telecom provider and AI provider.
- Cloud: managed operations, billed through Polar (Free/Developer/Pro/Scale).
- Enterprise: separate commercial terms and license for added extensions,
  supported deployment packages, SLAs and enterprise services.
- Paid Enterprise capabilities may be sold for Cloud or Self-Hosted without
  creating a second Voice Agent or telephony architecture.

See [enterprise-capabilities.md](enterprise-capabilities.md) and
[monetization.md](monetization.md).
