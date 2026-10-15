# Polar subscriptions and billing

Checkout and customer portal sessions use Polar. Configure `POLAR_ACCESS_TOKEN`
and `POLAR_WEBHOOK_SECRET` in the application environment. `POLAR_ENVIRONMENT`
defaults to `production`; use `sandbox` with sandbox credentials and product IDs
for acceptance testing. The API contract is pinned to `2026-10`.

The production product mappings default to:

| Plan | Polar product ID |
| --- | --- |
| Developer | `58ff657a-74d9-4342-89c4-2a3b784c6965` |
| Pro | `da5b3818-e92e-4b09-afad-ae61895f05db` |
| Scale | `a3023020-9822-4835-9509-1635c7d35f63` |

Override these with `POLAR_DEVELOPER_PRODUCT_ID`, `POLAR_PRO_PRODUCT_ID`, and
`POLAR_SCALE_PRODUCT_ID` when using another Polar environment. Configure recurring
products and portal upgrade/downgrade settings in Polar; local plan quotas remain
in Leamout. Checkout binds Polar's external customer ID to the Leamout organization
ID and maps the purchased product ID to a plan, independently of checkout metadata.

Configure a Polar webhook endpoint at `/v1/webhooks/polar` with API version
`2026-10` and subscription lifecycle events: created, updated, active, canceled,
uncanceled, cycled, revoked, past_due, paused, resumed, and migrated. The raw body
is verified against the webhook identity, timestamp and signature headers. Enable
`customer.state_changed` as well for immediate Customer State recovery. Both
new Standard Webhooks secrets and the documented legacy Polar signing scheme are
supported. Verified events are deduplicated by `webhook-id` in PostgreSQL before
Leamout responds, then a worker processes them with leases, exponential retries,
and dead-lettering. Repeated subscription notifications upsert the same local row,
and older observations cannot overwrite newer ones. Scheduled cancellation
preserves access until the current paid period ends. Customer State reconciliation
also checks organizations without a local subscription, allowing a missed initial
purchase event to be recovered.

When Polar reports `past_due`, Leamout keeps the current paid limits for seven days
by default. Set `POLAR_PAYMENT_GRACE_DAYS` to a value from 0 through 21 to change
this policy. The deadline is persisted when the delinquency begins, so webhook
retries and restarts cannot extend it. New calls use Free limits after the deadline,
while calls that are already active can finish.

The existing authenticated `/v1/billing/checkout` and `/v1/billing/portal` APIs
return Polar URLs. `/v1/billing/status` provides the console's plan, usage,
renewal, cancellation, and payment-grace state. After checkout, the console uses
`/v1/billing/reconcile` for bounded Customer State recovery while the webhook
worker catches up. Checkout still requires `Idempotency-Key` and uses Leamout's
local request-replay middleware. `cancel_url` is passed as Polar's `return_url`
for the checkout back button; success and return URLs must use the configured
frontend origin. No undocumented provider idempotency header is assumed.

Migration 034 creates the subscription schema with Polar as the supported billing
provider. Leamout enforces organization-level subscription limits and
runtime usage quotas for new agent sessions and call admission.

