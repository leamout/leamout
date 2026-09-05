# Rating

Rating converts metered usage quantities into commercial monetary amounts.

## Flow

```text
usage_event
    ↓
subscription
    ↓
plan
    ↓
usage rate resolution
    ↓
rated charge
    ↓
invoice_item snapshot
```

## Usage rates

The `usage_rates` model defines Leamout customer-facing billable usage pricing.

A rate can vary by:

```text
plan
meter
carrier provider
direction
country
network
currency
effective period
```

Fields include:

```text
plan_id
meter_id
carrier_provider_id
direction
country_code
network
currency
unit_amount_micros
unit_size
effective_from
effective_until
active
```

`carrier_provider_id` is an optional rating dimension that references the provider definition, not an organization's specific carrier connection. It lets Leamout select different customer pricing when the same usage is served by different upstream providers.

Usage rates are not upstream carrier costs. Actual completed-call provider costs are recorded separately from provider CDRs as `wholesale_charges`.

```text
usage_rates
    = what Leamout uses to price customer usage

provider_cdrs
    = what the upstream provider reports happened

wholesale_charges
    = what that provider usage actually cost Leamout
```

## Rate resolution

Current resolution requires:

```text
active, non-deleted organization
        ↓
subscription belongs to organization
        ↓
subscription status = active or past_due
        ↓
active plan and product
        ↓
active meter
        ↓
active effective usage rate
```

Optional rate dimensions behave as fallbacks. A null rate dimension matches any request value; a non-null dimension must match the requested value.

Current ranking prefers the candidate with the greatest number of specified dimensions across:

```text
carrier_provider_id
direction
country_code
network
```

then prefers the latest `effective_from`.

Conceptually:

```text
more specific match
        ↓
less specific match
        ↓
default rate
```

## Known resolution limitation

Specificity is currently based on the count of non-null dimensions. Two different candidates can therefore have equal specificity while matching different dimension combinations.

The schema also does not yet prevent overlapping effective periods for equivalent match keys.

Before large-scale production rating, define deterministic precedence and overlap policy explicitly rather than relying on insertion order.

## Money precision

Telecom prices can be far below one currency minor unit, so rates use micro-units:

```text
unit_amount_micros BIGINT
unit_size BIGINT
```

Example:

```text
unit_size = 60 seconds
unit_amount_micros = 80000
```

This can represent a charge of `0.08` currency units per 60 seconds when one currency unit equals 1,000,000 micros.

Invoice and payment totals use integer currency minor units. The rating service must use one deterministic rounding/conversion policy when converting rated micros into invoice amounts.

## Billing increments

The current schema does not yet model telecom-specific minimum/increment rules such as:

```text
60/60
6/6
1/1
```

Possible future fields include minimum and increment quantities. Add them only with a defined rating algorithm and tests.

## Historical pricing

A rated invoice item snapshots the rate/amount used at billing time. Historical invoices must not be recomputed from whatever `usage_rates` rows happen to be active later.
