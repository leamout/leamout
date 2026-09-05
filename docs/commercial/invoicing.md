# Invoicing

Invoicing turns fixed and rated usage charges into a durable commercial document whose historical amounts do not change when current pricing changes.

## Period flow

```text
billing period closes
        ↓
select billable usage
        ↓
resolve applicable rates
        ↓
aggregate and rate usage
        ↓
create draft invoice
        ↓
create invoice_items
        ↓
finalize/open invoice
        ↓
collect/reconcile payment
```

## Invoice

Current invoice fields include:

```text
organization_id
subscription_id
invoice_number
currency
subtotal
tax
total
status
issued_at
due_at
paid_at
metadata
created_at
updated_at
```

Current states are:

```text
draft
open
paid
void
overdue
```

The database validates the state names and basic amount/time relationships. Service logic should own valid transition rules.

## Invoice items

An invoice item is either:

```text
fixed
usage
```

A fixed line can represent a platform/subscription charge.

A usage line requires a meter and unit price and may reference the exact usage rate used.

Important fields are:

```text
invoice_id
meter_id
usage_rate_id
type
description
quantity
unit_amount_micros
amount
period_start
period_end
metadata
```

`amount` is stored in invoice currency minor units. `unit_amount_micros` preserves pricing precision for usage lines.

## Draft mutability

Current query behavior only creates or deletes invoice items while the invoice is in `draft` state.

This establishes a useful boundary:

```text
draft
  → composition may change

open/paid/void/overdue
  → historical lines should not be casually rewritten
```

Finalization behavior should be implemented as an explicit service operation.

## Historical price snapshot

Do not calculate an old invoice by looking up today's usage rate.

```text
usage at billing time
    ↓
usage rate selected at billing time
    ↓
invoice item stores rate reference + unit price + amount
```

Changing or deactivating a usage rate later must not change a finalized historical invoice.

## Tenant safety

Invoices are organization-owned. Invoice items derive tenant ownership through their invoice.

Creating or reading invoice items must verify:

```text
organization active/not deleted
        ↓
invoice belongs to organization
        ↓
item belongs to invoice
```

When a usage line references a usage rate and an invoice has a subscription, the rate should correspond to the subscription plan.

## Usage allocation is not solved yet

The current schema does not directly link usage events to invoice items. A production billing period needs a robust way to prove which usage has already been billed and to support corrections without double invoicing.

Do not solve this with a simplistic mutable `usage_events.invoiced = true` flag.

Potential future models include:

```text
usage_ratings
invoice_item_usage_events
billing_period_closures
```

Choose the model when period-close and correction behavior is implemented.

## Tax

The invoice currently contains an integer `tax` amount, but Leamout does not currently implement tax calculation/remittance infrastructure. Tax policy belongs to the billing/provider/application layer that actually owns that behavior.
