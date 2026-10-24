# Live Kinde identity acceptance

This suite checks real Kinde staging infrastructure without changing Leamout's
authentication model. Console and backoffice **share one Kinde tenant and one
confidential OAuth web application**. Both callback URLs and logout return URLs
must be allowlisted in that same Kinde application.

## Staging smoke test (manual GitHub Actions)

Create a GitHub Actions environment named `kinde-staging` with approval
requirements if appropriate. Configure:

| Type | Name | Value |
| --- | --- | --- |
| Secret | `KINDE_STAGING_ISSUER` | Full staging HTTPS issuer |
| Secret | `KINDE_STAGING_CLIENT_ID` | Shared confidential application ID |
| Secret | `KINDE_STAGING_CLIENT_SECRET` | Shared confidential application secret |
| Variable | `KINDE_STAGING_CONSOLE_CALLBACK_URL` | Registered console HTTPS callback |
| Variable | `KINDE_STAGING_BACKOFFICE_CALLBACK_URL` | Registered backoffice HTTPS callback |
| Optional variable | `KINDE_STAGING_CONSOLE_URL` | Deployed staging console HTTPS origin |
| Optional variable | `KINDE_STAGING_BACKOFFICE_URL` | Deployed staging backoffice HTTPS origin |

If you configure deployed origins, **configure both**. Values under
`KINDE_STAGING_*` are used only by the manually dispatched
`.github/workflows/kinde-acceptance.yml` workflow. They are not injected into
the ordinary Server, Docker, worker, media or acceptance workflows.

Run **Actions → Kinde live identity acceptance → Run workflow**.

The staging smoke verifies actual HTTPS OIDC discovery, issuer, OAuth endpoint
drift, reachable JWKS keys, verifier construction, login/registration request
parameters, PKCE/nonce/offline scope and both callback/logout return URLs.
If staging console and backoffice origins are given, it also verifies their
login redirects, state-cookie properties and unauthenticated account endpoints.

**Passing smoke tests is not proof that interactive login works.** These checks
do not register a person, consume a code, assert database membership, change
MFA settings or use provider refresh tokens. Configure *staging* hosts, not
production data or credentials.

## Interactive authentication-code acceptance (local only)

Use a disposable staging test user and a real interactive Kinde login to obtain
a **fresh, single-use** authorization code. Keep the matching PKCE verifier,
nonce and registered callback URL from the exact same authorization attempt.
The authorization code is bound to that callback and normally expires quickly.

Export the same issuer, client ID and client secret as above under the
`KINDE_LIVE_*` variable names, as well as:

- `KINDE_LIVE_CONSOLE_CALLBACK_URL` and `KINDE_LIVE_BACKOFFICE_CALLBACK_URL`
- `KINDE_LIVE_AUTH_CODE`, `KINDE_LIVE_PKCE_VERIFIER`, `KINDE_LIVE_NONCE`
- Optionally `KINDE_LIVE_TEST_REFRESH=true` to consume/rotate the refresh token
- Optionally `KINDE_LIVE_TEST_PORTAL=true` and
  `KINDE_LIVE_PORTAL_RETURN_URL` to exercise the account portal API

From `server/` run:

```sh
go test -tags=kinde_live -run '^TestLiveKindeAuthorizationCodeExchange$' \
  -count=1 -timeout=2m ./internal/integrations/kinde
```

The test requires a real code, verifies the token's signature, issuer, audience,
authorized party, nonce, subject and verified email via the production Go client,
and checks that an opaque session deadline can be established. Optional refresh
verifies the hosted token rotation path, while optional portal checks the
user self-service endpoint. The test deliberately does **not** print or persist
tokens, subjects, emails or portal URLs. Never store the code or refresh token in
GitHub secrets or logs. Never retry the same one-time code.

For local smoke without consuming an authorization code:

```sh
go test -tags=kinde_live -run '^TestLiveKinde' -count=1 \
  -timeout=5m ./internal/integrations/kinde
```

## Browser and Leamout database checks

For a release sign-off, record results **without PII or credentials**:

- Console registration, email verification, sign-in and logout
- Console callback accepts the correct state and rejects replay/mismatch
- Session stored under the verified Kinde issuer and subject
- Existing Leamout user UUID and organization memberships remain unchanged
- Console profile management redirects to Kinde then returns correctly
- Backoffice uses the same Kinde application but its own callback and cookie
- Non-platform-admin Kinde identity cannot enter backoffice admin routes
- Local session revoke and user-disable deny subsequent Leamout requests
- Expired/revoked refresh tokens deny access and require a new login
- Configured Kinde MFA rules are enforced by the hosted interactive login

The suite does not silently equate Kinde SSO logout with Leamout-wide logout,
and it does not claim provider-driven session revocation or automatic migration
from previous identity providers. Those need separate security work.
