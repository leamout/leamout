# Kinde authentication

Leamout uses **Kinde only** for hosted identity. Kinde owns sign-in, sign-up, email verification, MFA, password recovery and identity credentials. Leamout retains PostgreSQL users, organizations, memberships, billing, scoped API tokens, administrator authorization, and opaque browser sessions.

## Set up Kinde

1. Create a Kinde **regular web application** with a confidential client ID and secret.
2. Configure the issuer (for example `https://your-tenant.kinde.com`) and enable the Authorization Code flow, email and profile scopes, and refresh tokens (the Kinde `offline` scope).
3. Register **both** callback URLs:
   - `https://console.leamout.com/api/v1/auth/callback`
   - `https://backoffice.leamout.com/auth/callback`
4. Register `https://console.leamout.com/log-in` and `https://backoffice.leamout.com/log-in` as allowed logout return URLs.
5. Set `KINDE_ISSUER`, `KINDE_CLIENT_ID` and `KINDE_CLIENT_SECRET` in `server/.env`. Compose allowlists these keys to API and backoffice only. Do **not** add secrets to Next.js or browser bundles.
6. Require verified email in Kinde. Leamout rejects an unverified email. Configure Kinde's registration, authentication and MFA policies.

The Go backend performs OIDC authorization code + PKCE (S256), state cookie binding, nonce checking, ID token verification against OIDC discovery/JWKS, and encrypted refresh token storage. The frontend login and logout API remain `/api/v1/auth/*`; the standalone backoffice uses `/auth/*`.

## Session and migration behavior

Identity is keyed by `(identity_issuer, identity_subject)`; Leamout refuses to silently merge users with matching email addresses. **An existing Keycloak user will not automatically transfer to Kinde.** Provision identities in Kinde first and prepare a reviewed mapping of Kinde subjects to the corresponding local UUIDs before switching traffic. Keep organization memberships and subscription references untouched. Revoke old browser sessions when cutting over. Never delete an existing Keycloak database or volume until user identity data and any rollback requirements have been backed up and checked.

Kinde can omit a new ID token during refresh. Leamout retains the prior verified subject and validates any newly returned ID token. Provider login/logout is a browser action; local browser sessions are explicitly revoked. No provider back-channel logout support is assumed. Revoking an identity in Kinde may not immediately revoke a cached local Leamout session; revoke it locally when disabling access.

Existing database migrations and SQLC-generated models are deliberately preserved for upgrade compatibility; removing unused tables is separate reviewed migration work. Leamout does not proxy Kinde under `auth.leamout.com`. Remove the old `auth` DNS record only after the cutover and rollback window.
