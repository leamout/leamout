# Catalog

The Catalog defines Leamout-owned customer-facing commercial terms. It is global configuration: catalog records do not belong to an organization and do not contain wallet, payment, license, deployment, or usage state.

## Model

```text
product
   ↓
 plan
   ↓
price

meter ─────┘  (for metered prices)
```

A **product** is a commercial product family. A **plan** groups reusable offers within that product. A **price** is one immutable set of customer-facing terms. A **meter** identifies a measurable quantity that a metered price can reference.

## Pricing types

Current price types are:

```text
one_time
recurring
metered
```

A one-time price uses `amount_minor`.

A recurring price uses `amount_minor` plus a billing interval such as `month` or `year`.

A metered price uses:

```text
meter_id
unit_amount_micros
unit_size
dimensions
```

Recurring prices do not imply a customer subscription lifecycle. They can describe recurring product charges, such as number rental or self-hosted licensing terms, while the current Cloud commercial model remains prepaid PAYG.

## Product and plan identity

Product and plan codes are stable machine identifiers used to resolve offers without coupling callers to database UUIDs.

Catalog records can be retired from new use without rewriting historical customer authorizations or commercial records.

## Customer price versus provider cost

Catalog prices answer what Leamout charges the customer.

They do not represent DIDWW, CommPeak, or other provider wholesale costs.

```text
Catalog price
    = customer-facing Leamout price

provider wholesale charge
    = Leamout COGS
```

Provider inventory prices, SKUs, rate sheets, and CDR cost must never become the authoritative customer price merely because the provider fulfilled the operation.

## PAYG use

For a fixed managed purchase, the application resolves the applicable Leamout Catalog offer before wallet authorization.

```text
customer selects managed service
        ↓
resolve Leamout Catalog price
        ↓
reserve wallet funds
        ↓
create provider obligation
        ↓
capture or release
```

For metered products, the Catalog can represent customer-facing unit terms. The system must still preserve the terms used for an authorization so later Catalog changes do not rewrite an already-authorized operation.

## Current application boundary

`commercial/catalog` is read-oriented application behavior. It resolves and lists configured products, plans, prices, and meters used by concrete commercial workflows.

Catalog mutation should be added only with a concrete trusted operator/configuration workflow.

## Boundaries

The Catalog answers:

```text
What does Leamout sell?
What offer/plan identifies it?
What customer-facing price applies?
What meter/unit defines a metered price?
Is this commercial term currently available?
```

It does not answer:

```text
How much prepaid balance does an organization have?
Did a provider collect money?
Is a self-hosted deployment licensed?
What did DIDWW or CommPeak charge Leamout?
```

Those belong to Wallets, Payments, Licensing, and wholesale/provider reconciliation respectively.
