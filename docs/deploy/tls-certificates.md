# TLS certificates

Leamout uses TLS certificates for SIP TLS in OpenSIPS.

The deployment expects certificate material under:

```text
deploy/certs/
├── fullchain.pem
├── privkey.pem
└── carrier-ca.pem
```

These files are mounted into the OpenSIPS container by `deploy/compose.yaml`.

Do not commit certificate private keys or production certificate material to the repository. `deploy/certs/*.pem` is ignored by Git.

## Certificate roles

The three files have different responsibilities:

| File | Purpose | Production source |
| --- | --- | --- |
| `fullchain.pem` | OpenSIPS server certificate and certificate chain | Let's Encrypt / Certbot |
| `privkey.pem` | Private key for `fullchain.pem` | Let's Encrypt / Certbot |
| `carrier-ca.pem` | CA trust bundle used to validate outbound SIP carrier TLS peers | System CA bundle or carrier-provided CA bundle |

`carrier-ca.pem` is not the Leamout server certificate. It is used by the OpenSIPS outbound carrier TLS client configuration to verify the remote carrier certificate chain.

## Local development and CI

For local development and CI, generate a self-signed certificate:

```bash
make certs
```

This runs:

```text
scripts/certs/generate-self-signed.sh
```

and creates the required files under `deploy/certs/`.

The self-signed generator refuses to overwrite an existing certificate set unless replacement is explicitly requested with `CERT_FORCE=1`.

Example:

```bash
CERT_FORCE=1 make certs
```

Self-signed certificates are intended for development, automated tests, and controlled private environments. Use a publicly trusted certificate for an Internet-facing production deployment.

Validate the current certificate set with:

```bash
make check-certs
```

The validation checks that:

- all required files exist;
- the server certificate is valid PEM;
- the private key is valid;
- the certificate and private key match;
- the server certificate has not expired; and
- `carrier-ca.pem` contains readable certificate material.

Both `make up` and `make deploy` run `make check-certs` before starting the stack.

## Production with Let's Encrypt

Leamout supports production certificate provisioning with Certbot and Let's Encrypt.

### 1. Choose a SIP TLS hostname

Use a hostname dedicated to the SIP endpoint, for example:

```text
sip.example.com
```

Create a DNS `A` and/or `AAAA` record that points the hostname to the public IP address of the Leamout VPS.

The hostname used for SIP TLS must match the hostname on the Let's Encrypt certificate.

### 2. Allow the ACME HTTP challenge

The production provisioning script uses Certbot's standalone HTTP challenge.

Before requesting the certificate:

- DNS for the hostname must resolve to the VPS;
- inbound TCP port `80` must be allowed by the VPS firewall and cloud/provider firewall; and
- no other process may occupy TCP port `80` while Certbot performs the challenge.

SIP TLS itself continues to use the SIP TLS listener configured by Leamout. Port `80` is needed only for the HTTP ACME challenge used by this Certbot configuration.

### 3. Install Certbot

Install Certbot using the package supported by the VPS operating system.

For example, on a Debian/Ubuntu host where the `certbot` package is available:

```bash
sudo apt update
sudo apt install certbot
```

Verify it is available:

```bash
certbot --version
```

### 4. Configure Leamout

Create the production environment file before deployment:

```bash
cp .env.example .env
nano .env
```

Replace all development placeholders and secrets with production values.

The production deployment intentionally uses `.env`; `.env.example` is only a template.

### 5. Request the production certificate

Run:

```bash
sudo TLS_DOMAIN=sip.example.com \
  TLS_EMAIL=admin@example.com \
  make certs-production
```

`TLS_DOMAIN` is required and must be the public SIP hostname.

`TLS_EMAIL` is required for the Let's Encrypt account registration and certificate notifications.

The command runs:

```text
scripts/certs/provision-letsencrypt.sh
```

The script:

1. verifies that Certbot is installed;
2. requests a certificate with Certbot standalone mode if one does not already exist;
3. reads the Certbot-managed certificate from `/etc/letsencrypt/live/<domain>/`;
4. copies `fullchain.pem` and `privkey.pem` into `deploy/certs/` with appropriate permissions;
5. installs a carrier CA bundle if `carrier-ca.pem` does not already exist; and
6. validates the resulting Leamout certificate set.

Certbot remains the source of truth for the Let's Encrypt certificate. The files under `deploy/certs/` are the runtime copies mounted into OpenSIPS.

### 6. Deploy

After certificate provisioning succeeds:

```bash
make deploy
```

A typical first production deployment is therefore:

```bash
git clone https://github.com/leamout/leamout.git
cd leamout

cp .env.example .env
nano .env

sudo apt update
sudo apt install certbot

sudo TLS_DOMAIN=sip.example.com \
  TLS_EMAIL=admin@example.com \
  make certs-production

make deploy
```

## Carrier CA trust

Outbound SIP carrier TLS is configured to verify the carrier certificate against:

```text
deploy/certs/carrier-ca.pem
```

When `make certs-production` runs and `carrier-ca.pem` does not already exist, Leamout copies a supported system CA bundle from the VPS.

This is appropriate when the SIP carrier presents a certificate issued by a normal publicly trusted CA.

If a carrier supplies its own CA certificate or private CA bundle, install that instead:

```bash
sudo CARRIER_CA_FILE=/path/to/carrier-ca.pem \
  sh scripts/certs/install-system-ca.sh
```

The installer does not overwrite an existing `deploy/certs/carrier-ca.pem`. Remove or replace that file deliberately when changing the carrier trust configuration.

After changing carrier trust material, validate it:

```bash
make check-certs
```

and restart OpenSIPS:

```bash
docker compose --env-file .env -f deploy/compose.yaml restart opensips
```

## Renewal

Let's Encrypt certificates are short-lived and must be renewed periodically.

Leamout provides:

```bash
sudo TLS_DOMAIN=sip.example.com make certs-renew
```

This runs:

```text
scripts/certs/renew-letsencrypt.sh
```

The renewal helper:

1. runs `certbot renew` for the configured certificate name;
2. copies the current Certbot `fullchain.pem` and `privkey.pem` into `deploy/certs/`;
3. validates the resulting certificate set; and
4. restarts the OpenSIPS container so it loads the current certificate.

The renewal command still requires the ACME challenge method saved by Certbot to succeed. With the current standalone configuration, TCP port `80` must therefore remain usable when Certbot actually performs a renewal challenge.

### Automating renewal

Certificate renewal should be scheduled on production hosts rather than relying on an operator to remember it.

Many Certbot packages install their own renewal timer. However, Leamout also needs the renewed files copied into `deploy/certs/` and OpenSIPS restarted afterward.

Until a Leamout-specific systemd timer or Certbot deploy hook is installed, schedule the Leamout renewal helper itself.

For example, a root cron entry can periodically run:

```cron
17 3 * * * cd /opt/leamout && TLS_DOMAIN=sip.example.com make certs-renew >> /var/log/leamout-cert-renew.log 2>&1
```

Running the command daily is safe because Certbot renews only when the managed certificate is close enough to expiry.

Use the real repository path instead of `/opt/leamout`.

## Custom certificate directory

The default runtime directory is:

```text
deploy/certs
```

The certificate scripts and Makefile support overriding it with `CERT_DIR`:

```bash
make CERT_DIR=/srv/leamout/certs check-certs
```

If the Docker Compose mounts are not also changed to use the same directory, OpenSIPS will continue to mount `deploy/certs`. For the standard deployment, keep the default directory.

## Troubleshooting

### `TLS_DOMAIN is required`

Provide the public SIP hostname:

```bash
sudo TLS_DOMAIN=sip.example.com \
  TLS_EMAIL=admin@example.com \
  make certs-production
```

### Certbot cannot complete the HTTP challenge

Check that:

- the hostname resolves to the correct VPS public IP;
- TCP port `80` is open in all firewalls/security groups;
- NAT forwards port `80` to the VPS when applicable; and
- another service is not listening on port `80` during the standalone challenge.

### `Missing required certificate file`

Inspect:

```bash
ls -la deploy/certs
```

For development/CI:

```bash
make certs
```

For production:

```bash
sudo TLS_DOMAIN=sip.example.com \
  TLS_EMAIL=admin@example.com \
  make certs-production
```

### Certificate and private key do not match

Do not mix certificate files from different Certbot certificate names or certificate issuances.

Re-sync the correct Certbot-managed certificate with:

```bash
sudo TLS_DOMAIN=sip.example.com \
  TLS_EMAIL=admin@example.com \
  make certs-production
```

### Carrier TLS verification fails

Verify whether the carrier uses a publicly trusted CA or supplies a carrier-specific CA bundle.

If a carrier-specific bundle is required, install the carrier-provided CA as `deploy/certs/carrier-ca.pem`, run `make check-certs`, and restart OpenSIPS.

## Command reference

```text
make certs
    Generate self-signed certificates for development and CI.

make certs-production
    Request or install the Let's Encrypt production server certificate.
    Requires TLS_DOMAIN and TLS_EMAIL and must run as root.

make certs-renew
    Run Certbot renewal, sync the current certificate into deploy/certs,
    validate it, and restart OpenSIPS. Requires TLS_DOMAIN and root.

make check-certs
    Validate the runtime certificate files.

make deploy
    Validate certificates, pull the latest main branch, build/start the
    production Compose stack, and show service status.
```
