# Wallets

Wallets are Leamout's prepaid monetary boundary for all Cloud and managed communications charges. PostgreSQL is the source of truth for wallet value; Redis may coordinate realtime admission but never creates or settles money.

Self-Hosted software licenses are the sole commercial exception to the wallet-funded PAYG model. They are enterprise software agreements and are settled separately.

## Core records

```text
wallets
wallet_ledger_entries
wallet_reservations
```

Each wallet belongs to one organization and one currency.

A wallet does not represent a subscription, enterprise license, provider account, or telecom product.

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
verified bank-transfer funding
refund
credit adjustment
```

Examples of debits:

```text
captured managed-service charge
captured Cloud/platform charge
chargeback
debit adjustment
```

A correction is a new compensating entry. Existing posted ledger rows are not rewritten.

Enterprise Self-Hosted license settlement must not be posted as wallet credit merely because the customer paid Leamout by bank transfer.

## Reservation flow

```text
chargeable operation requested
        ↓
lock wallet
        ↓
verify spendable balance
        ↓
create reservation when required
        ↓
commit
        ↓
charge/provider exposure allowed
        ↓
operation succeeds → capture
operation fails before obligation → release
```

For managed-provider operations, provider exposure must never happen before committed prepaid authorization.

For other Cloud/platform consumption, the service must likewise prevent chargeable use from exceeding authorized prepaid value.

## Concurrency

Reservation admission serializes on the wallet. Two concurrent operations cannot both authorize the same remaining balance.

Capture and release are lifecycle operations on an existing reservation and must remain retry-safe.

For long-running managed operations such as voice, additional authorization must be committed before extending the provider-cost horizon. A target-total reservation model is preferred over retry-sensitive delta increases.

## Customer pricing

Wallets store and move money. They do not define customer prices.

Customer-facing price comes from Catalog or from a concrete service-specific pricing decision made before authorization.

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

BYOC means the customer owns the carrier relationship. That carrier's charges do not consume a Leamout managed-carrier wallet merely because traffic is controlled by Leamout.

This does not create a postpaid Leamout model. Any chargeable Leamout Cloud/platform consumption still uses prepaid PAYG.

For Self-Hosted + BYOC, the enterprise Self-Hosted software license is the relevant Leamout commercial obligation unless the customer also uses a Leamout managed service.

## Self-Hosted + Managed

A Self-Hosted + Managed customer has two separate commercial relationships:

```text
enterprise Self-Hosted software license
        +
prepaid managed-usage wallet
```

The license does not grant managed-usage credit. Managed usage must be funded and authorized from the wallet before provider exposure.

## Non-goals

Wallets do not own:

- customer subscriptions;
- enterprise Self-Hosted contract negotiation;
- customer pricing definitions;
- telecom fulfillment;
- provider wholesale costing;
- postpaid credit;
- customer withdrawals;
- FX conversion;
- invoice-centric usage settlement.
