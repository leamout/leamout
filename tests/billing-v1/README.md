# Polar billing acceptance

This suite exercises Leamout's billing HTTP handlers, durable PostgreSQL webhook
inbox, subscription reconciliation, and effective limits. It uses a local fake
Polar API and signed synthetic Polar-format events. It does not claim that real
Polar checkout, charges, or renewals have been tested.

## Run the contract acceptance suite

```sh
sh tests/billing-v1/run.sh
```

The runner needs Go matching `server/go.mod` and Docker. It creates a disposable
PostgreSQL 17 container on a random loopback port and removes it afterward. It
does not start the voice stack or use your deployment's environment file.

Alternatively, set `BILLING_TEST_DATABASE_URL` to a dedicated acceptance database.
The Go harness creates an isolated schema and removes it when the test finishes.
Use a test database only. Existing PostgreSQL installations need schema-creation
permission. CI uses a dedicated PostgreSQL service.

The contract scenarios include initial purchase, renewal, upgrade, downgrade,
scheduled cancellation, immediate revocation, failed-payment grace and recovery,
duplicate delivery, durable retry, and recovery when the first webhook is missed.
Assertions use the billing HTTP endpoints and effective limits, with database
checks for durable inbox processing.

## Optional read-only sandbox checks

Polar sandbox and production have separate tokens and product IDs. Configure
the sandbox products explicitly; the runner never defaults to production IDs.
Provide secrets through your secret manager or CI environment rather than
putting tokens in shell history or committing them.

Required environment variables:

- `POLAR_SANDBOX_ACCESS_TOKEN`, with `products:read` and `customers:read` scopes.
- `POLAR_SANDBOX_DEVELOPER_PRODUCT_ID`.
- `POLAR_SANDBOX_PRO_PRODUCT_ID`.
- `POLAR_SANDBOX_SCALE_PRODUCT_ID`.

Optionally set `POLAR_SANDBOX_CUSTOMER_EXTERNAL_ID` to the UUID of a Leamout
organization that has already used sandbox checkout. The script reads Customer
State and verifies that Polar retained this external customer identity.

```sh
python3 tests/billing-v1/sandbox.py
```

Set `BILLING_RUN_POLAR_SANDBOX_CHECKS=1` to run this check after the contract suite.
It only sends authenticated GET requests to `https://sandbox-api.polar.sh`,
refuses redirects, and does not display credentials or customer response bodies.
It checks recurring monthly prices of $79, $299, and $799 in USD. It does not
create checkout sessions, customers, subscriptions, payments, or webhook deliveries.

## Remaining sandbox lifecycle acceptance

Run these steps against a dedicated Leamout Cloud test installation with
`POLAR_ENVIRONMENT=sandbox`, sandbox product IDs, a sandbox access token, and a
sandbox webhook secret. Register `/v1/webhooks/polar` with Polar, select the
subscription events and `customer.state_changed`, and use only the payment
methods documented in [Polar sandbox testing](https://polar.sh/docs/integrate/sandbox).
Never enter a real payment method into these tests.

| Scenario | Action in Polar sandbox | Expected Leamout result |
| --- | --- | --- |
| Purchase | Complete Developer checkout from the billing page. | Checkout redirects successfully; paid plan and Developer limits appear after asynchronous processing. |
| Renewal | Use Polar's available sandbox renewal tools, or wait for the sandbox billing cycle. | The subscription retains paid access and reports the provider's new period dates. |
| Upgrade | Change Developer to Pro through the customer portal. | The effective plan and capacity match the provider's effective change; no duplicate subscription is created. |
| Downgrade | Change Pro to Developer through the portal. | Capacity changes when the provider makes the downgrade effective. A scheduled change must not apply early. |
| Scheduled cancellation | Cancel at the current period's end. | Cancellation is displayed while paid access remains through the paid period; limits fall back after expiry. |
| Immediate revocation | Revoke the test subscription immediately. | Paid access ends when Leamout processes the authoritative revocation; active calls follow existing drain behavior. |
| Failed payment | Use Polar's documented payment-failure simulation. | Status becomes `past_due`; paid limits remain for Leamout's configured grace period, then fall back. |
| Payment recovery | Resolve the simulated failure before and after grace expiry. | Active status restores paid limits and clears the failure/grace timestamps. |
| Duplicate event | Redeliver a previously accepted webhook from Polar's delivery tools. | Delivery receives an acknowledgment; the durable inbox has one event per `webhook-id`; usage and periods are unchanged. |
| Worker outage | Stop the billing worker, complete a purchase, then restart it. | Webhooks are acknowledged only after durable storage; the worker applies the queued event after restart. |
| Missed first webhook | Complete a new checkout while delivery is unavailable, then restore it. | Customer State reconciliation discovers the subscription without requiring a local subscription row. |

Confirm the billing page's plan, limits, usage, renewal/cancellation status,
grace-period notice, and portal link after each step. Record the sandbox event
types and observed results in the PR; keep raw payloads containing personal data
and credentials out of Git. Passing local contract tests and read-only sandbox
checks does not complete this manual payment lifecycle checklist.
