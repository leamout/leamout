# Catalog

The Catalog defines Leamout-owned customer-facing commercial terms. It is global configuration: catalog records do not belong to an organization and do not contain wallet, payment, license, deployment, or usage state.

Everything in Leamout is prepaid pay-as-you-go except Self-Hosted software licenses. Catalog pricing must preserve that distinction.

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

Recurring catalog terms do not create a customer subscription lifecycle and do not permit postpaid usage. A recurring term can describe a product that renews periodically, such as number rental, while each charge remains prepaid.

Self-Hosted software license terms may also be represented in Catalog for offer/pricing configuration, but enterprise license settlement is handled separately from the PAYG wallet path.

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

For any chargeable Cloud or managed operation, the applicable Leamout customer term is resolved before prepaid authorization.

For a fixed managed purchase:

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

For metered products, the Catalog can represent customer-facing unit terms. The system must preserve the terms used for an authorization so later Catalog changes do not rewrite an already-authorized operation.

Cloud + BYOC remains prepaid PAYG for any Leamout platform charge even though the customer's carrier cost is outside Leamout.

## Self-Hosted license exception

Self-Hosted software licensing is the sole non-PAYG commercial path.

Catalog may describe the enterprise offer, but payment and license issuance are not wallet-funded checkout flows.

```text
Catalog license offer
        ↓
enterprise agreement / invoice
        ↓
verified settlement
        ↓
license lifecycle
```

Paying for a Self-Hosted license does not create wallet balance or managed-usage credit.

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
