# Security model

Leamout treats organization isolation, deployment sovereignty, and telecom fraud
controls as separate security boundaries. A control in one boundary must not be
used as a substitute for a control in another.

## Leamout Cloud

The organization selected by an authenticated request is the tenant boundary.
Organization tokens carry their organization identity; session requests must
select an organization and prove current membership. Repositories must include
the organization in reads, updates, and deletes, and writes that reference a
parent resource must prove that the parent belongs to the same organization.
Globally unique resource IDs are not authorization credentials.

Carrier credentials are encrypted before persistence. New carrier credential
ciphertexts use AES-GCM associated data containing the organization, carrier
connection, and credential direction. Moving ciphertext to another tenant,
connection, or credential field therefore makes authentication fail. Provider
master credentials are platform configuration: customer APIs and tenant data
models must never return or accept them.

The customer-facing carrier service deliberately receives encryption authority
without decryption authority. Digest credential rotation replaces the secret and
its realm-bound HA1 material; it does not read the previous plaintext. Credential
export is not supported. If a future carrier protocol requires the original
secret at execution time, decryption belongs in a separate least-privilege
runtime component that is scoped to that organization, connection, and purpose.

Cloud deployments should additionally use a dedicated secret manager or KMS,
private database and cache networks, separate workload identities, encrypted
backups, and database credentials that do not permit schema-owner bypass during
normal request handling. Logs, traces, metrics labels, audit metadata, and error
responses must not contain credentials or call content.

## Self-hosted deployments

A self-hosted runtime must be able to operate without exporting customer call
data. Database, Redis, NATS, SIP, media, recordings, backups, logs, metrics, and
traces remain inside the customer's network unless the operator explicitly
configures an external destination. Licensing and update checks must exchange
only deployment and entitlement data; they must not include tenant resources,
destinations, recordings, message bodies, or carrier credentials.

Production installations must:

- generate unique encryption, TURN, internal API, database, NATS, Redis, and
  FreeSWITCH secrets rather than retaining example values;
- keep secret files readable only by the service identity and keep them out of
  images, source control, command lines, support bundles, and logs;
- expose only documented public SIP, media, TURN, and HTTPS ports, leaving
  PostgreSQL, Redis, NATS, FreeSWITCH ESL, metrics, and internal admission APIs
  on private networks;
- terminate HTTPS with a valid certificate, use host firewall allowlists, and
  restrict administrative access through a VPN or equivalent trusted network;
- verify signed release manifests before upgrades, run containers without
  unnecessary privileges, pin images by digest, and regularly test encrypted
  backup restoration; and
- rotate credentials after installation or suspected compromise and revoke
  enrollment credentials after they have been exchanged.

An installation that enables an external telemetry exporter or stores backups
outside the VPC is no longer operating in strict no-egress mode. That choice
must be explicit and documented by the operator.

## Telecom and infrastructure

All call origination must pass admission before a carrier is contacted. The
admission boundary is fail closed and uses shared state so adding API or worker
replicas cannot multiply limits. Enforce, at minimum, calls per second,
concurrent calls, and daily carrier usage. Daily spend caps, destination and
country allowlists, premium-rate blocking, verified caller IDs, and anomaly
alerts should be configured to the customer's risk profile before enabling
PSTN traffic.

Never fall back from an explicitly selected BYOC route to a managed carrier.
Inbound tenant identity must be derived from an assigned number or another
trusted routing record, never from a caller-controlled SIP header. Internal SIP
admission and reconciliation endpoints require independent machine
credentials, private network reachability, bounded request bodies, replay
protection where messages can be retried, and credential rotation.

Rate limiting is required at both the HTTP edge and telecom admission layer.
HTTP limits must key on the server-resolved organization and credential—not an
organization header supplied by the caller. Authentication, authorization,
quota, billing reservation, and routing failures must fail closed.

## Security invariants for changes

Every tenant-owned feature should include tests proving that a valid principal
from another organization cannot read, mutate, enumerate, or infer the
resource. Secret-bearing features should include tests for redaction,
ciphertext tampering, and cross-context ciphertext replay. Telecom changes
should test quota exhaustion, shared-state failure, lease release, destination
policy, and the absence of route fallback.

Security-sensitive events—authentication changes, credential lifecycle,
membership and role changes, carrier configuration, number reassignment, and
quota overrides—must produce immutable audit records containing actor,
organization, target, action, and time, but never plaintext secrets.
