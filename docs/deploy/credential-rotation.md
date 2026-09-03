# Certificate and shared-secret rotation

Leamout treats deployment credentials as four different rotation classes. They do not share one safe cutover mechanism:

| Material | Consumers | Rotation model |
| --- | --- | --- |
| TLS `fullchain.pem` / `privkey.pem` / `carrier-ca.pem` | OpenSIPS, Coturn | stage, validate, drain, activate, reload/restart, resume |
| `TURN_AUTH_SECRET` | API credential issuer, Coturn | old/new overlap for the full short-lived credential lifetime |
| `FREESWITCH_ESL_PASSWORD` | FreeSWITCH, API, worker | drain and coordinated process recreation |
| `CARRIER_CREDENTIAL_ENCRYPTION_KEY` | API and encrypted carrier credential rows | transactional database re-encryption before changing the process key |

Never rotate one of these values by editing `.env` and restarting arbitrary containers. The commands below encode the required ordering.

Production secrets remain in `.env` or an external secret-management system. Do not commit them to Git.

## TLS certificate rotation

For a manually supplied certificate set, place the new files in a staging directory outside `deploy/certs` and run:

```sh
sudo SOURCE_CERT_DIR=/secure/new-leamout-certs make rotate-certs
```

The source directory must contain `fullchain.pem` and `privkey.pem`. If it also contains `carrier-ca.pem`, the carrier trust bundle is rotated with the server certificate. Otherwise the currently active carrier CA bundle is retained.

The activation workflow:

1. copies the candidate files into a temporary staging directory;
2. validates certificate readability, expiry, certificate/private-key match, and CA readability;
3. gracefully drains OpenSIPS and FreeSWITCH and waits for active sessions to finish;
4. backs up the current runtime files;
5. overwrites the existing bind-mounted files in place so running containers do not retain stale mount inodes;
6. validates the activated runtime set;
7. sends `SIGUSR2` to Coturn so its TLS context reloads the new certificate; and
8. starts/resumes RTPengine, FreeSWITCH, and OpenSIPS.

If activation fails after the runtime files were changed, the script restores the previous runtime certificate files and leaves the command failed for operator inspection.

### Let's Encrypt

The Certbot deploy hook uses the same activation path. The existing production flow remains:

```sh
sudo TLS_DOMAIN=sip.example.com make certs-renew
sudo TLS_DOMAIN=sip.example.com make certs-auto-renew
```

A successful Certbot renewal therefore receives the same validation and graceful activation behavior as a manual certificate replacement.

### Verify

After rotation:

```sh
make check-certs
make verify
```

Also verify SIP TLS/WSS and TURN TLS from an external client when those listeners are Internet-facing.

## TURN shared-secret rotation

TURN REST credentials issued by Leamout are valid for 10 minutes. Replacing the shared secret immediately would invalidate credentials that were issued before the cutover but have not yet expired.

Coturn supports multiple static TURN REST secrets. Leamout uses `TURN_AUTH_SECRET_PREVIOUS` during a controlled overlap window so both old and new credentials remain valid while the API begins issuing only the new secret.

Generate a new secret and begin rotation:

```sh
export NEW_TURN_AUTH_SECRET="$(openssl rand -hex 32)"
make rotate-turn-secret
unset NEW_TURN_AUTH_SECRET
```

The begin phase:

1. rejects a nested rotation if `TURN_AUTH_SECRET_PREVIOUS` is already populated;
2. drains active calls before recreating Coturn;
3. moves the current secret to `TURN_AUTH_SECRET_PREVIOUS` in `.env`;
4. installs the new value as `TURN_AUTH_SECRET`;
5. records the rotation start time;
6. recreates Coturn first so it accepts both secrets;
7. recreates the API so new ICE credentials are signed only with the new secret; and
8. resumes call admission.

Wait at least 660 seconds by default. The extra minute is safety margin beyond the 10-minute API credential lifetime. Then finalize:

```sh
make finalize-turn-secret
```

Finalization refuses to remove the previous secret before the overlap window has elapsed. It drains active calls, clears the previous secret, recreates Coturn, and resumes admission.

`TURN_ROTATION_MIN_OVERLAP_SECONDS` may be increased, but values below 600 are rejected. `TURN_ROTATION_FORCE=1` exists only for emergency invalidation of the previous secret; using it can invalidate still-live credentials.

## FreeSWITCH ESL password rotation

Generate a new strong password and run:

```sh
export NEW_FREESWITCH_ESL_PASSWORD="$(openssl rand -hex 32)"
make rotate-esl-password
unset NEW_FREESWITCH_ESL_PASSWORD
```

The rotation drains active calls, updates `.env`, recreates FreeSWITCH/API/worker together, verifies FreeSWITCH ESL authentication and API readiness with the new password, then resumes telecom admission.

If readiness does not recover, the node stays drained. Do not resume admission until FreeSWITCH, API, and worker all agree on the same password.

## Carrier credential encryption-key rotation

`CARRIER_CREDENTIAL_ENCRYPTION_KEY` encrypts the inbound and outbound carrier credential ciphertext stored in PostgreSQL. Changing only the environment variable makes existing rows undecryptable.

Generate a base64url-encoded 32-byte AES key:

```sh
export NEW_CARRIER_CREDENTIAL_ENCRYPTION_KEY="$(openssl rand -base64 32 | tr '+/' '-_' | tr -d '=')"
make rotate-carrier-key
unset NEW_CARRIER_CREDENTIAL_ENCRYPTION_KEY
```

The command:

1. builds the server image containing `/leamout/rotate-carrier-key` before downtime;
2. stops API and worker so no carrier credential writes race the migration;
3. locks every carrier connection containing encrypted credential material in a serializable transaction;
4. decrypts each ciphertext with the old key and re-encrypts it with the new key;
5. accepts rows already encrypted with the new key so a partially completed operator workflow is safely rerunnable;
6. rolls back the whole database transaction if any ciphertext decrypts with neither key;
7. updates `.env` only after the database transaction commits;
8. starts API and worker with the new key; and
9. verifies API readiness.

If the database transaction fails, API and worker are restarted with the old key. If the transaction commits but environment activation fails, the script deliberately does **not** restart the API with the old key. Re-run the command with the same new key to finish activation.

### Backup requirement

Take a database backup before rotating the carrier credential encryption key. The command is transactional, but the backup is the recovery boundary for operator mistakes such as supplying the wrong historical key.

## Emergency compromise response

If a secret is believed compromised, prioritize revocation over graceful overlap:

- TLS private key compromise: activate a newly issued certificate/key immediately and revoke the old certificate with the issuing CA where supported.
- TURN secret compromise: begin a normal rotation; if continued acceptance of the old secret is unacceptable, finalize with `TURN_ROTATION_FORCE=1` and accept that outstanding credentials may fail.
- ESL password compromise: run the coordinated ESL rotation immediately; active calls may need to be drained or terminated according to incident severity.
- Carrier encryption-key compromise: rotate the key and re-encrypt the database immediately, then rotate the underlying carrier digest secrets themselves because disclosure of the encryption key may have exposed recoverable plaintext.

## Rotation cadence

Use certificate validity/issuer automation for TLS. For shared secrets and encryption keys, define an organization policy based on incident response and secret-management requirements rather than rotating so frequently that controlled maintenance becomes unsafe. Always test the rotation commands in staging after changes to Compose, OpenSIPS, Coturn, FreeSWITCH, or the credential schema.
