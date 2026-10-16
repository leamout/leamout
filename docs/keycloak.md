# Keycloak authentication

Leamout Cloud uses one Keycloak realm for user authentication. Keycloak owns credentials, registration, email verification, password recovery, MFA, and identity sessions. Leamout retains organization membership, resource authorization, billing, and organization API tokens. Keycloak realm roles never grant Leamout organization or platform administrator access.

## Deployment

1. Point `auth.leamout.com` at the Caddy edge. Set `DOMAIN=leamout.com` in the application environment. If using another domain, update `KC_HOSTNAME` in Compose and `KEYCLOAK_ISSUER` together.
2. Set the Keycloak database and bootstrap administrator passwords in root `.env`. Set `KEYCLOAK_CLIENT_SECRET` to a generated secret in `server/.env`.
3. Set `FRONTEND_URL` to the console's exact public origin and `BACKEND_URL` to the API's public origin. The API must be able to reach the public Keycloak issuer. Keycloak must be able to reach the backchannel logout URL.
4. Start the cloud Compose stack. It includes a dedicated Keycloak PostgreSQL database, Keycloak, and an `auth` reverse proxy route. Keycloak imports `deploy/keycloak/leamout-realm.json` only when the realm does not already exist.
5. Configure the realm's SMTP settings, using SES SMTP credentials if desired. SES SMTP credentials are separate from the API credentials used by Leamout's invitation worker. Email verification is required, so registration cannot complete until SMTP works.
6. Replace the bootstrap administrator with an operational administrator and remove bootstrap credentials from the deployment. Back up the Keycloak database separately from Leamout.

The imported client is confidential, uses Authorization Code + S256 PKCE, and disables direct password grants, implicit grants, and service accounts. It registers exactly:

- Callback: `<FRONTEND_URL>/api/v1/auth/callback`
- Post-logout return: `<FRONTEND_URL>/log-in`
- Backchannel logout: `<BACKEND_URL>/v1/identity/backchannel-logout`

The Next.js rewrite must preserve `Set-Cookie` and proxy the callback to `/v1/auth/callback`. No client secret or provider token is exposed to Next.js client components. Changes to the realm file do not update an existing realm: apply subsequent client, URL, secret, or policy changes through Keycloak administration.

For development, use a local Keycloak issuer such as `http://localhost:8180/realms/leamout`, `APP_ENV=development`, and the console origin `http://localhost:3001`. Configure the same exact callback and logout paths in that development realm. Production requires HTTPS. Local development is for Leamout contributors, not a separate hosting edition.

## Identity and session behavior

The API verifies ID token signatures, issuer, audience, expiration, nonce, and any access-token hash. A nonempty subject, session ID, and verified email are required. The local user is keyed by `(identity_issuer, identity_subject)`; email collisions fail rather than linking identities. Disabled local users cannot sign in or refresh.

Login state is bound to an HttpOnly, SameSite=Lax browser cookie and a five-minute, single-use database record. PKCE verifiers are encrypted. Sessions use opaque, hashed browser tokens; provider tokens are encrypted using `ENCRYPTION_KEY` and bound to the local user. Cookies are host-only and secure in production.

The local session expires after at most 30 days, and Keycloak's idle/max lifetime still applies. Before the verified tokens expire, requests use the cached session. At expiration, a database row lock serializes refresh-token rotation across API replicas. A failed refresh rejects authentication. A crash after provider rotation but before the database commit can require signing in again.

Logout clears the local session and sends the browser to Keycloak's end-session endpoint. Signed backchannel logout events revoke matching local sessions, with token-ID replay protection. Account profile, password, MFA, and device/session settings are managed in Keycloak's account console. The remaining `/sessions` endpoints manage Leamout's local application sessions only; they do not terminate SSO in other applications.

Keycloak is outside the audio pipeline. Its outage can block login and refresh, but it does not terminate established voice calls. If a backchannel notification cannot be delivered, authentication stops when the cached token expires and refresh fails.

## Schema and rollout

This change rewrites the unapplied base migrations and regenerates SQLC/Atlas outputs. It removes the old credential/OTP tables and custom directory/federation schema. It is not an in-place migration for an already populated deployment.

Before deploying against an existing database, prepare a reviewed data migration that maps each existing local user ID to a verified Keycloak issuer and subject, preserves organization references, and invalidates old browser sessions. Do not automatically link by email or apply rewritten migration history to an existing database.

Leamout's SES worker remains for organization invitations; Keycloak sends authentication emails.
