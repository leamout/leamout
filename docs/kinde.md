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

## Live acceptance

Before production rollout, follow [Live Kinde identity acceptance](./kinde-acceptance.md). The opt-in staging workflow covers live OIDC metadata and login entry points, while the single-use local authorization-code test and interactive browser checks validate actual user authentication, sessions, refresh and account management.
