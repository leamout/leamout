# TLS certificates

Leamout uses the same deployment-managed server certificate for OpenSIPS SIP TLS/WSS and Coturn TURN TLS. Outbound SIP carrier TLS also uses the deployment-managed carrier CA bundle.

Runtime material lives under:

```text
deploy/certs/
├── fullchain.pem
├── privkey.pem
└── carrier-ca.pem
```

Do not commit private keys or production certificate material. `deploy/certs/*.pem` is ignored by Git.

| File | Purpose | Typical production source |
| --- | --- | --- |
| `fullchain.pem` | OpenSIPS and Coturn server certificate/chain | Let's Encrypt / Certbot |
| `privkey.pem` | private key for `fullchain.pem` | Let's Encrypt / Certbot |
| `carrier-ca.pem` | trust bundle for outbound SIP carrier TLS | system CA or carrier-provided CA |

## Development and CI

Generate a self-signed set:

```sh
make certs
```

The generator refuses to overwrite an existing set unless replacement is explicit:

```sh
CERT_FORCE=1 make certs
```

Validate runtime material with:

```sh
make check-certs
```

Validation checks required files, certificate/private-key matching, expiry, and readable CA material. `make up` and `make deploy` run certificate validation before starting the stack.

## Production with Let's Encrypt

Choose a public SIP/TURN hostname, for example `sip.example.com`, and point DNS to the Leamout host. The current provisioning path uses Certbot's standalone HTTP challenge, so TCP port 80 must be reachable when Certbot performs the challenge.

Install Certbot using the host operating system, configure `.env`, then request/install the certificate:

```sh
sudo TLS_DOMAIN=sip.example.com \
  TLS_EMAIL=admin@example.com \
  make certs-production
```

Certbot remains the source of truth under `/etc/letsencrypt/live/<domain>/`; `deploy/certs/` contains runtime copies mounted into Leamout services.

Install automatic renewal:

```sh
sudo TLS_DOMAIN=sip.example.com make certs-auto-renew
```

This installs the Leamout deploy hook under Certbot's renewal hooks and enables `certbot.timer` when the host package provides it.

## Safe activation after renewal

Certificate renewal and certificate activation are separate steps. Leamout's deploy hook routes renewed material through `scripts/certs/activate-runtime.sh` instead of blindly replacing files and restarting one service.

Activation:

1. stages the candidate certificate/key and current or replacement carrier CA;
2. validates the staged set before touching runtime files;
3. gracefully drains current calls;
4. backs up the active runtime files;
5. overwrites the existing bind-mounted files in place;
6. validates the activated runtime set;
7. sends `SIGUSR2` to Coturn to reload its TLS context; and
8. starts/resumes RTPengine, FreeSWITCH, and OpenSIPS, causing OpenSIPS' script-defined TLS domains to load the new files.

OpenSIPS' `tls_mgm:reload` reloads database-defined TLS domains; Leamout currently defines its TLS domains in `opensips.cfg`, so certificate-file replacement requires OpenSIPS process startup/restart rather than relying on that MI command.

Manual renewal uses the same activation path:

```sh
sudo TLS_DOMAIN=sip.example.com make certs-renew
```

For a certificate supplied outside Certbot, stage it in another directory and activate it directly:

```sh
sudo SOURCE_CERT_DIR=/secure/new-leamout-certs make rotate-certs
```

See `docs/deploy/credential-rotation.md` for rollback behavior, secret rotation, and incident-response procedures.

## Carrier CA trust

`carrier-ca.pem` validates outbound SIP carrier certificates. It is not Leamout's server certificate.

When `make certs-production` runs and no carrier CA exists, Leamout can install a supported system CA bundle. If a carrier supplies a private CA bundle, install it deliberately:

```sh
sudo CARRIER_CA_FILE=/path/to/carrier-ca.pem \
  sh scripts/certs/install-system-ca.sh
```

To rotate the carrier CA with the server certificate, include `carrier-ca.pem` in `SOURCE_CERT_DIR` before running `make rotate-certs`. If the staging directory omits it, activation retains the current carrier CA bundle.

## Testing renewal

Verify Certbot's ACME renewal configuration separately:

```sh
sudo certbot renew --dry-run
```

Inspect scheduling:

```sh
systemctl status certbot.timer
systemctl list-timers --all | grep certbot
```

After a real or staged activation:

```sh
make check-certs
make verify
```

Also test SIP TLS/WSS and TURN TLS from an external client when those listeners are public.

## Custom certificate directory

Scripts support `CERT_DIR`:

```sh
make CERT_DIR=/srv/leamout/certs check-certs
```

The standard Compose file still mounts `deploy/certs`. If `CERT_DIR` is changed for a production deployment, the Compose mounts must point to the same runtime directory.

## Troubleshooting

### `TLS_DOMAIN is required`

Supply the public hostname:

```sh
sudo TLS_DOMAIN=sip.example.com \
  TLS_EMAIL=admin@example.com \
  make certs-production
```

### ACME HTTP challenge fails

Confirm DNS points to the correct host, TCP 80 is open through all firewalls/NAT layers, and no other process occupies port 80 while standalone Certbot is performing a challenge.

### Certificate and private key do not match

Do not mix files from different issuances. Re-sync the correct Certbot lineage or stage a matching pair, then run `make check-certs` before activation.

### Rotation drains but does not resume

The activation workflow intentionally leaves a failed deployment in a conservative state. Inspect `docker compose --env-file .env -f deploy/compose.yaml ps`, validate certificates, correct the problem, and run:

```sh
ENV_FILE=.env COMPOSE_FILE=deploy/compose.yaml sh deploy/resume.sh
```

### Carrier TLS verification fails

Determine whether the carrier uses a public CA or a private carrier-specific CA. Validate the correct bundle as `carrier-ca.pem` and activate it through the certificate rotation workflow.

## Command reference

```text
make certs
    Generate self-signed development/CI certificates.

make certs-production
    Request/install the initial Let's Encrypt certificate.

make certs-auto-renew
    Install the Certbot deploy hook and renewal timer integration.

make certs-renew
    Run Certbot renewal and safely activate the resulting certificate.

make rotate-certs
    Validate and activate SOURCE_CERT_DIR through graceful drain/reload.

make check-certs
    Validate the current runtime certificate set.
```
