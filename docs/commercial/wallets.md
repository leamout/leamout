# Wallets

Wallets are Leamout's prepaid monetary boundary. PostgreSQL is the source of truth for wallet value; Redis may coordinate realtime admission but never creates or settles money.

## Core records

```text
wallets
wallet_ledger_entries
wallet_reservations
```

Each wallet belongs to one organization and one currency.

A wallet does not represent a subscription, plan assignment, provider account, or telecom product.

## Posted and spendable balance

Ledger entries are immutable posted credits and debits.

Spendable balance is:

```text
posted credits
- posted debits
- active reservations
```

Reservations temporarily hold prepaid value before an operation is captured or released.

## Ledger rules

Ledger history is append-only.

Examples of credits:

```text
wallet top-up
refund
credit adjustment
```

Examples of debits:

```text
captured managed-service charge
chargeback
debit adjustment
```

A correction is a new compensating entry. Existing posted ledger rows are not rewritten.

## Reservation flow

```text
managed operation requested
        ↓
lock wallet
        ↓
verify spendable balance
        ↓
create reservation
        ↓
commit
        ↓
provider exposure allowed
        ↓
operation succeeds → capture
operation fails before obligation → release
```

Provider exposure must never happen before the reservation commit that authorizes it.

## Concurrency

Reservation admission serializes on the wallet. Two concurrent operations cannot both authorize the same remaining balance.

Capture and release are lifecycle operations on an existing reservation and must remain retry-safe.

For long-running managed operations such as voice, additional authorization must be committed before extending the provider-cost horizon. A target-total reservation model is preferred over retry-sensitive delta increases.

## Customer pricing

Wallets store and move money. They do not define customer prices.

Customer-facing price comes from Catalog or from a concrete service-specific pricing decision made before reservation.

Provider wholesale cost remains separate:

```text
Leamout customer price
        ↓
wallet authorization

DIDWW / CommPeak cost
        ↓
wholesale COGS reconciliation
```

A provider-reported cost must not be used as authority to debit the customer wallet.

## Managed number purchase

The prepaid sequence for a managed DID is:

```text
resolve Leamout customer price
        ↓
reserve matching-currency wallet funds
        ↓
create provider order
        ↓
provider fails/cancels → release
provider completes      → capture
```

Insufficient funds must fail before the provider order is created.

Managed number provider configuration must not enable automatic renewal that can create future provider obligations without a new Leamout authorization.

## Managed voice

Managed voice requires rolling prepaid authorization because final duration is not known at call start.

```text
reserve initial authorized horizon
        ↓
originate managed call
        ↓
raise target reservation before extending horizon
        ↓
stop extension when funds are unavailable
        ↓
call ends
        ↓
capture final customer amount
```

The final customer charge must use Leamout-authorized customer terms, not a provider CDR cost. Provider CDRs remain a separate wholesale reconciliation concern.

## BYOC

BYOC carrier usage does not consume Leamout managed-carrier wallet value merely because traffic is controlled by Leamout.

Commercial observation and provider-cost authorization are separate concerns.

## Non-goals

Wallets do not own:

- subscriptions;
- entitlements/access policy;
- customer pricing definitions;
- telecom fulfillment;
- provider wholesale costing;
- postpaid credit;
- customer withdrawals;
- FX conversion;
- invoice-centric settlement.
