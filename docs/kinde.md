# Kinde authentication

Leamout uses **Kinde only** for hosted identity. Kinde owns sign-in, sign-up, email verification, MFA, password recovery and identity credentials. Leamout retains PostgreSQL users, organizations, memberships, billing, scoped API tokens, administrator authorization, and opaque browser sessions.

## Set up Kinde

1. Create **one Kinde regular web application** with a confidential client ID and secret. Both the Leamout console and backoffice use this same Kinde application and tenant.
2. Configure the issuer (for example `https://your-tenant.kinde.com`) and enable the Authorization Code flow, email and profile scopes, and refresh tokens (the Kinde `offline` scope).
3. Register **both** callback URLs in the same Kinde application:
   - `https://console.leamout.com/api/v1/auth/callback`
   - `https://backoffice.leamout.com/auth/callback`
4. Register `https://console.leamout.com/log-in` and `https://backoffice.leamout.com/log-in` as allowed logout return URLs.
5. Set `KINDE_ISSUER`, `KINDE_CLIENT_ID` and `KINDE_CLIENT_SECRET` in `server/.env`. Compose allowlists these same credentials to the API and backoffice only. Do **not** add secrets to Next.js or browser bundles.
6. Require verified email in Kinde. Leamout rejects an unverified email. Configure Kinde's registration, authentication and MFA policies.

Both entry points share Kinde's sign-in system, but use their own callback URLs and host-only Leamout session cookies. Backoffice permissions are enforced separately through Leamout's platform-administrator authorization; a valid Kinde login alone does not grant operator access.

The Go backend performs OIDC authorization code + PKCE (S256), state cookie binding, nonce checking, ID token verification against OIDC discovery/JWKS, and encrypted refresh token storage. The frontend login and logout API remain `/api/v1/auth/*`; the standalone backoffice uses `/auth/*`.

## Session and migration behavior

Identity is keyed by `(identity_issuer, identity_subject)`; Leamout refuses to silently merge users with matching email addresses. **An existing Keycloak user will not automatically transfer to Kinde.** Provision identities in Kinde first and prepare a reviewed mapping of Kinde subjects to the corresponding local UUIDs before switching traffic. Keep organization memberships and subscription references untouched. Revoke old browser sessions when cutting over. Never delete an existing Keycloak database or volume until user identity data and any rollback requirements have been backed up and checked.

Kinde can omit a new ID token during refresh. Leamout retains the prior verified subject and validates any newly returned ID token. Provider login/logout is a browser action; local browser sessions are explicitly revoked. No provider back-channel logout support is assumed. Revoking an identity in Kinde may not immediately revoke a cached local Leamout session; revoke it locally when disabling access.

Existing database migrations and SQLC-generated models are deliberately preserved for upgrade compatibility; removing unused tables is separate reviewed migration work. Leamout does not proxy Kinde under `auth.leamout.com`. Remove the old `auth` DNS record only after the cutover and rollback window.

## Refresh and audience security

The verifier enforces signature, issuer and expiration using OIDC discovery/JWKS. It then accepts a standard client-ID audience, or the Kinde issuer audience only when the `azp` claim matches the configured confidential client; other audiences fail closed. Leamout revalidates Kinde refresh-token authority at least every ten minutes while a session is active, even when Kinde issues longer-lived access tokens. A failed refresh rejects access. The single Kinde OAuth application must allowlist both callback URLs and both logout return URLs.

## Self-service account management

Leamout uses Kinde's **user** self-service portal for personal identity/profile changes. From Console Settings > Profile or Security, an authenticated browser requests `GET /api/v1/auth/account`, which checks the host-only Leamout session, refreshes the encrypted Kinde token as needed, calls `GET /account_api/v1/portal_link?subnav=profile&return_url=...` using a server-side bearer access token, and redirects the user to the one-time Kinde portal link. The access token never reaches Next.js. The destination must remain an HTTPS URL on the configured Kinde issuer origin. The Backoffice Account link uses the same Kinde tenant and OAuth client but returns to the backoffice and requires Leamout platform-admin authorization.

Kinde's user portal must be enabled in the Kinde environment. Password recovery, authentication factor policies, and account security must be configured and tested in Kinde; Leamout does not implement a parallel password or MFA database. Console Security also lists/revokes **Leamout-local** browser sessions; revoking those sessions does not revoke all Kinde SSO sessions. Kinde organizations, billing, and API keys must not be used as substitutes for Leamout-owned data.

## Synchronizing Kinde identity changes

Leamout receives signed Kinde `user.updated` and `user.deleted` webhooks at:

`POST https://api.leamout.com/v1/identity/kinde/webhook`

Configure this HTTPS address in **Kinde → Settings → Webhooks** and subscribe only to those two event types. Console and backoffice continue sharing a single Kinde identity provider and application.

Kinde delivers `Content-Type: application/jwt` with a signed JWT body and a `webhook-id` header. The Go integration retrieves Kinde's published `/.well-known/jwks` and requires a valid RS256 signature and signing-key ID. The event must have a valid `event_id`, timestamp within the accepted replay window, and a matching `kp_` or `kp:` subject.

Each accepted event is recorded once in `kinde_identity_events`, keyed by provider issuer and event ID. A signed `user.updated` event updates the **existing** Leamout user's profile display name and email, avoiding email collisions with other local users. It cannot create a user, change a user UUID, membership, platform-administrator status, or restore a disabled user. Out-of-order older profile updates do not overwrite newer updates.

A signed `user.deleted` event or a user update indicating suspension marks the matching Leamout user disabled and revokes existing browser sessions. The event's durable disable marker also prevents a Kinde subject deleted before first sign-in from being provisioned later. Revocation of all existing local browser sessions is completed in the same database statement that records the signed event; unrelated identities remain unaffected. That is a conservative, irreversible local-access decision. Restoring access requires an explicit reviewed administrator action; Kinde webhook updates cannot silently reactivate accounts. Unknown Kinde user subjects do not create local accounts.

Webhook verification failure returns a non-success status; transient JWKS/database errors produce an HTTP 503 so Kinde can retry. Production rollout requires setting up the provider webhook, applying migration `036_create_kinde_identity_events.sql`, regenerating sqlc/Atlas artifacts, and exercising signed provider events in staging. Kinde-originated changes to email or credentials must be verified against a real tenant; the local account link is always the verified `(identity_issuer, identity_subject)`, **never email alone**.

This endpoint must be excluded from any browser-origin restrictions intended only for session-backed frontend requests. Do not put bearer tokens, webhook JWTs, or user PII in application logs.
