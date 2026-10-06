# Transactional email with AWS SES v2

The API persists authentication challenges and encrypted email jobs in one PostgreSQL transaction. The worker renders embedded HTML and text templates and sends with AWS SDK for Go v2 / SES v2. Email jobs are private database records, not public outbox events.

## Configuration

Apply migration `032_create_email_deliveries.sql` before starting the API server and worker. Set these variables on the API server and worker:

```env
AWS_REGION=eu-west-1
EMAIL_FROM=Leamout <notifications@your-domain.com>
SES_CONFIGURATION_SET=leamout-transactional
```

`SES_CONFIGURATION_SET` is optional. The worker needs AWS credentials from the SDK default credential chain; prefer an IAM role. The API only queues jobs and does not need AWS permissions. Both processes must share the existing `ENCRYPTION_KEY`. Email delivery is always enabled. Startup requires a non-empty `AWS_REGION` and a valid `EMAIL_FROM` address.

Verify the sending domain and DKIM in the configured SES region, configure MAIL FROM/SPF and DMARC, and request production access when leaving the sandbox. Give the worker `ses:SendEmail` scoped to the verified identity. If a configuration set is specified, create it in the same region. Live AWS setup and sending are separate from local tests.

## Templates

`internal/platform/email/templates` contains a shared HTML layout and HTML/text templates for OTP and invitations. Templates are embedded at build time. HTML uses automatic escaping. Invitation URLs must use HTTPS. Invitation rendering and queueing are available; the existing invitation endpoints remain unimplemented.

Queue an email using `email.Service.QueueTx` with the business operation's `pgx.Tx`. Payloads use scoped authenticated encryption bound to the delivery ID, and are erased after acceptance, terminal failure, expiry, or supersession. Rotate encryption keys only with a plan for pending jobs and existing encrypted credentials.

## Delivery behavior

- Workers claim one job at a time using `FOR UPDATE SKIP LOCKED`, a 60-second lease, and a unique lease token.
- Each provider request has a 15-second timeout and SDK automatic retries are disabled. The queue retries temporary errors, up to five attempts, with exponential delays.
- Deliveries have an optional opaque `cancellation_key`, with no foreign key to authentication or other business tables. The queue worker checks only delivery status, lease ownership, and expiry.
- Auth owns OTP lifecycle rules. Resend invalidates earlier challenges and cancels matching email jobs in the same transaction. Successful OTP/password login and exhausted OTP attempts cancel queued jobs and erase their payloads. Expiry follows the authentication transaction's original ten-minute deadline.
- All delivery persistence queries live in `internal/database/queries/email_deliveries.sql`; repositories use generated sqlc methods. Queueing and cancellation use the caller's transaction.
- Resend cooldown is 60 seconds; each recipient can request five codes per hour. Public auth endpoints have a shared Redis limit of 30 requests/minute per socket peer. Configure per-client limits at your reverse proxy too; forwarded headers are not trusted for this budget.
- SES message IDs record provider acceptance, not inbox delivery. A crash or uncertain timeout can produce duplicate emails. Revocation concurrent with an in-flight provider request cannot retract that email; its code is invalidated.
- Raw provider errors and email bodies are never stored in error fields. Only classified error codes are persisted.

Configure SES delivery/bounce/complaint events and suppression monitoring in AWS before production rollout. This change supplies sending and delivery jobs; an application event-feedback ingestion endpoint is not included.

## Tests

```sh
go test -race ./internal/platform/email ./internal/integrations/ses ./internal/identity/auth ./internal/platform/middleware
EMAIL_TEST_DATABASE_URL=postgres://postgres:password@localhost:5432/testdb?sslmode=disable go test -race ./internal/identity/auth
```

The integration tests create isolated schemas, apply the relevant migrations, and drop their schemas afterward. Use a dedicated test database. No tests send real emails.
