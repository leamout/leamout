# Leamout Cloud — Private Beta

This is a single-node, assisted-onboarding launch. Reuse the existing Docker
Compose services; there is no Kubernetes or managed-carrier architecture.

## Required preparation

1. Recruit a small group of developers or agencies that already have their
   own SIP provider/PBX and AI account.
2. Choose a Linux host with public IPv4, Docker Engine/Compose, Python 3,
   OpenSSL and curl. Start with an estimated 4 vCPU/8 GB RAM and benchmark
   before making any concurrency guarantees.
3. Configure public DNS for api, recordings, and turn subdomains.
   The Next.js Console needs separate hosting and HTTPS; Caddy does not
   currently serve the Console.
4. Create private .env and server/.env from the example files (chmod 600).
   Replace all example secrets, align DATABASE_URL with POSTGRES_PASSWORD,
   and set APP_ENV=production, DOMAIN, PUBLIC_IP, FRONTEND_URL,
   BACKEND_URL=https://api.DOMAIN, CORS_ORIGINS, TURN_REALM, TURN_PUBLIC_URLS.
5. Provision TLS files at /etc/leamout/certs/fullchain.pem and
   /etc/leamout/certs/privkey.pem for SIP/TURN. Renew them separately from
   Caddy's ACME certificates. Provision a separate MinIO application user.
6. Configure transactional SES email and verify sign-up flows before inviting
   customers. Store tenant BYOA credentials in Leamout, not in env files.
7. Restrict SSH, SIP signaling, RTP, and TURN relay ingress at the cloud
   provider firewall. The base Compose publishes 23000–32768/UDP for RTP
   and 49152–65535/UDP for TURN; explicitly assess and restrict these ranges.
   Do not expose Postgres, Redis, NATS, FreeSWITCH ESL, worker, media, or
   MinIO internal ports to the public. Docker published ports can bypass
   naive UFW filtering; apply provider-side rules too.

## Start and validate

From the repository root:

    python3 deploy/cloud/preflight.py
    APP_ENV_FILE=../server/.env docker compose --env-file .env -f deploy/compose.yaml -f deploy/cloud/compose.yaml config --quiet
    APP_ENV_FILE=../server/.env docker compose --env-file .env -f deploy/compose.yaml -f deploy/cloud/compose.yaml up -d --build
    sh deploy/cloud/smoke.sh

Preflight checks production configuration, secret placeholders, environment
file permissions, mounted SIP/TURN certificate validity and Docker Compose
configuration. Smoke checks container and public API readiness.
Neither script provisions DNS/certificates nor places billable calls.

## Assisted customer onboarding

- Bootstrap an organization via supported authentication flows. Never use
  credentials seeded by tests. Validate invitation and SES email delivery.
- Connect one customer-owned SIP trunk, configure authorized signaling sources,
  and bind a customer-owned number.
- Connect and verify an organization-owned AI provider integration.
- Create, bind and activate a Voice Agent. Verify readiness and place one
  real consented inbound and outbound call. Check durable conversation,
  recording, usage and cleanup. PR #36 contains a separate opt-in live tester.
- Validate merchant eligibility, real subscription checkout, webhook
  synchronization and payout before accepting paying customers. Polar
  approval is not implied by dashboard access.

## Backups and operations

Set up automated encrypted off-host backups for PostgreSQL and recording
objects before onboarding customers. Back up encryption keys and necessary
certificates separately with strict access. Test restore into an isolated
deployment before announcing beta readiness. This PR does **not** implement
backup scheduling, disaster recovery, or high availability.

Keep an operator on call for the first pilots, inspect container logs and
readiness checks daily, and stop onboarding when calls, backups or email
verification fail. Test the first real deployment before public marketing.

## Explicit remaining work

No server has been provisioned by this PR. Real DNS, certificates, firewall
policy, operational backup/restore, Console hosting, first organization,
carrier, BYOA provider, live calls and payment merchant approval remain
deployment and business launch prerequisites.
